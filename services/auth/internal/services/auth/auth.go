package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/lib/jwt"
	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

const (
	verificationTTL = 10 * time.Minute
	resendCooldown  = 1 * time.Minute
	maxAttempts     = 5
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCode        = errors.New("invalid code")
	ErrInvalidToken       = errors.New("invalid token")
	ErrCodeCooldown       = errors.New("code resend cooldown not expired")
	ErrNoAttemptsLeft     = errors.New("no attempts left")
	ErrCodeNotFound       = errors.New("verification code not found")
)

type Auth struct {
	log                 *slog.Logger
	userSaver           UserSaver
	userProvider        UserProvider
	verificationStorage VerificationStorage
	jwtSecret           string
	accessTokenTTL      time.Duration
	refreshTokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (int, error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
}

type VerificationStorage interface {
	SaveVerificationData(ctx context.Context, email string, passHash []byte, codeHash string, attempts int, ttl time.Duration) error
	GetVerificationData(ctx context.Context, email string) (passHash []byte, codeHash string, attempts int, createdAt time.Time, err error)
	DecrementAttempts(ctx context.Context, email string) (int, error)
	DeleteVerificationData(ctx context.Context, email string) error
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	verificationStorage VerificationStorage,
	jwtSecret string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:                 log,
		userSaver:           userSaver,
		userProvider:        userProvider,
		verificationStorage: verificationStorage,
		jwtSecret:           jwtSecret,
		accessTokenTTL:      accessTokenTTL,
		refreshTokenTTL:     refreshTokenTTL,
	}
}

func generateCode() (string, error) {
	max := new(big.Int).SetInt64(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashCode(code string) string {
	h := sha256.Sum256([]byte(code))
	return hex.EncodeToString(h[:])
}

// SendCode initiates registration when a password is provided, or resends the
// verification code (with cooldown) when the password is empty.
func (a *Auth) SendCode(ctx context.Context, email string, password string) (string, error) {
	if password == "" {
		// Empty password means the client is requesting a resend.
		return a.resendCode(ctx, email)
	}

	const op = "auth.SendCode"

	log := a.log.With(slog.String("op", op))

	log.Info("sending verification code", slog.String("email", email))

	// Enforce resend cooldown if a verification is already pending.
	_, _, _, createdAt, err := a.verificationStorage.GetVerificationData(ctx, email)
	if err == nil {
		if time.Since(createdAt) < resendCooldown {
			log.Info("resend cooldown not expired")
			return "", fmt.Errorf("%s: %w", op, ErrCodeCooldown)
		}
		// Cooldown passed: replace old verification data.
		if err := a.verificationStorage.DeleteVerificationData(ctx, email); err != nil {
			log.Error("failed to delete old verification data", slog.String("error", err.Error()))
			return "", fmt.Errorf("%s: %w", op, err)
		}
	} else if !errors.Is(err, storage.ErrCodeNotFound) {
		log.Error("failed to get verification data", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Check if user already exists in DB
	_, err = a.userProvider.User(ctx, email)
	if err == nil {
		log.Info("user already exists")
		return "", fmt.Errorf("%s: %w", op, ErrUserExists)
	}
	if !errors.Is(err, storage.ErrUserNotFound) {
		log.Error("failed to check user existence", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Hash password
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Generate 6-digit code
	code, err := generateCode()
	if err != nil {
		log.Error("failed to generate code", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	codeHash := hashCode(code)

	// Save to Redis (overwrites any existing entry for this email)
	err = a.verificationStorage.SaveVerificationData(ctx, email, passHash, codeHash, maxAttempts, verificationTTL)
	if err != nil {
		log.Error("failed to save verification data", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("verification code sent")

	return code, nil
}

// resendCode generates a new code if the cooldown (1 min) has passed.
// Keeps the same password hash, resets attempts and creation time.
func (a *Auth) resendCode(ctx context.Context, email string) (string, error) {
	const op = "auth.ResendCode"

	log := a.log.With(slog.String("op", op))

	log.Info("resending verification code", slog.String("email", email))

	// Get existing verification data
	passHash, _, _, createdAt, err := a.verificationStorage.GetVerificationData(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrCodeNotFound) {
			log.Info("no pending verification found")
			return "", fmt.Errorf("%s: %w", op, ErrCodeNotFound)
		}
		log.Error("failed to get verification data", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Check cooldown (must wait at least 1 minute since last code)
	if time.Since(createdAt) < resendCooldown {
		log.Info("resend cooldown not expired")
		return "", fmt.Errorf("%s: %w", op, ErrCodeCooldown)
	}

	// Delete old entry
	if err := a.verificationStorage.DeleteVerificationData(ctx, email); err != nil {
		log.Error("failed to delete old verification data", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Generate new code
	code, err := generateCode()
	if err != nil {
		log.Error("failed to generate code", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	codeHash := hashCode(code)

	// Save new data with the same passHash
	err = a.verificationStorage.SaveVerificationData(ctx, email, passHash, codeHash, maxAttempts, verificationTTL)
	if err != nil {
		log.Error("failed to save new verification data", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("verification code resent")

	return code, nil
}

// Register verifies the 6-digit code and, if valid, persists the user in the DB
// and returns an access + refresh token pair.
func (a *Auth) Register(ctx context.Context, email string, code string) (string, string, error) {
	const op = "auth.Register"

	log := a.log.With(slog.String("op", op))

	log.Info("registering user", slog.String("email", email))

	// Get verification data from Redis
	passHash, storedCodeHash, attempts, _, err := a.verificationStorage.GetVerificationData(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrCodeNotFound) {
			log.Info("no pending verification found")
			return "", "", fmt.Errorf("%s: %w", op, ErrCodeNotFound)
		}
		log.Error("failed to get verification data", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	// Check that there are attempts remaining
	if attempts <= 0 {
		log.Info("no attempts left")
		_ = a.verificationStorage.DeleteVerificationData(ctx, email)
		return "", "", fmt.Errorf("%s: %w", op, ErrNoAttemptsLeft)
	}

	// Verify code hash
	if hashCode(code) != storedCodeHash {
		remaining, err := a.verificationStorage.DecrementAttempts(ctx, email)
		if err != nil {
			log.Error("failed to decrement attempts", slog.String("error", err.Error()))
			return "", "", fmt.Errorf("%s: %w", op, err)
		}

		if remaining <= 0 {
			log.Info("no attempts left, deleting verification data")
			_ = a.verificationStorage.DeleteVerificationData(ctx, email)
			return "", "", fmt.Errorf("%s: %w", op, ErrNoAttemptsLeft)
		}

		log.Info("invalid code", slog.Int("remaining_attempts", remaining))
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCode)
	}

	// Code is correct — save user to the database
	userID, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Info("user already exists")
			return "", "", fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save user", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	// Clean up verification data
	_ = a.verificationStorage.DeleteVerificationData(ctx, email)

	// Generate token pair
	user := models.User{
		ID:       userID,
		Email:    email,
		PassHash: passHash,
	}

	accessToken, err := jwt.NewToken(user, a.jwtSecret, a.accessTokenTTL, "access")
	if err != nil {
		log.Error("failed to create access token", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := jwt.NewToken(user, a.jwtSecret, a.refreshTokenTTL, "refresh")
	if err != nil {
		log.Error("failed to create refresh token", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered successfully", slog.Int("user_id", userID))

	return accessToken, refreshToken, nil
}

// Login authenticates a user and returns an access + refresh token pair.
func (a *Auth) Login(ctx context.Context, email string, password string) (string, string, error) {
	const op = "auth.Login"

	log := a.log.With(slog.String("op", op))

	log.Info("logging in user")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Info("user not found")
			return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to get user", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	// Compare password hashes without revealing which check failed.
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Info("invalid password")
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("user logged in")

	accessToken, err := jwt.NewToken(user, a.jwtSecret, a.accessTokenTTL, "access")
	if err != nil {
		log.Error("failed to create access token", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := jwt.NewToken(user, a.jwtSecret, a.refreshTokenTTL, "refresh")
	if err != nil {
		log.Error("failed to create refresh token", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, refreshToken, nil
}

// RefreshTokens validates a refresh token and issues a new token pair.
func (a *Auth) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	const op = "auth.RefreshTokens"

	log := a.log.With(slog.String("op", op))

	log.Info("refreshing tokens")

	// Parse and validate refresh token payload.
	claims, err := jwt.ParseToken(refreshToken, a.jwtSecret)
	if err != nil {
		log.Info("invalid refresh token", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		log.Info("token is not a refresh token")
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	// Ensure the token carries an email claim.
	email, ok := claims["email"].(string)
	if !ok {
		log.Info("email not found in token")
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Info("user not found")
			return "", "", fmt.Errorf("%s: %w", op, ErrInvalidToken)
		}
		log.Error("failed to get user", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	newAccessToken, err := jwt.NewToken(user, a.jwtSecret, a.accessTokenTTL, "access")
	if err != nil {
		log.Error("failed to create access token", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	newRefreshToken, err := jwt.NewToken(user, a.jwtSecret, a.refreshTokenTTL, "refresh")
	if err != nil {
		log.Error("failed to create refresh token", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("tokens refreshed")

	return newAccessToken, newRefreshToken, nil
}
