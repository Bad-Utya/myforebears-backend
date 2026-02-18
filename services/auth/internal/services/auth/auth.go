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
	log                  *slog.Logger
	userSaver            UserSaver
	userProvider         UserProvider
	verificationStorage  VerificationStorage
	jwtSecret            string
	accessTokenTTL       time.Duration
	refreshTokenTTL      time.Duration
	linkForResetPassword string
	linkTTL              time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (int, error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
}

type VerificationStorage interface {
	SaveCode(ctx context.Context, email string, passHash []byte, codeHash string, attempts int, ttl time.Duration) error
	GetCode(ctx context.Context, email string) (passHash []byte, codeHash string, attempts int, createdAt time.Time, err error)
	VerifyCode(ctx context.Context, email string, inputCodeHash string) (passHash []byte, matched bool, err error)
	SaveLink(ctx context.Context, email string, linkHash string, ttl time.Duration) error
	GetLink(ctx context.Context, linkHash string) (email string, err error)
	DeleteCode(ctx context.Context, email string) error
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	verificationStorage VerificationStorage,
	jwtSecret string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	linkForResetPassword string,
	linkTTL time.Duration,
) *Auth {
	return &Auth{
		log:                  log,
		userSaver:            userSaver,
		userProvider:         userProvider,
		verificationStorage:  verificationStorage,
		jwtSecret:            jwtSecret,
		accessTokenTTL:       accessTokenTTL,
		refreshTokenTTL:      refreshTokenTTL,
		linkForResetPassword: linkForResetPassword,
		linkTTL:              linkTTL,
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

func generateLink() (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[n.Int64()]
	}
	return string(b), nil
}

func hashString(str string) string {
	h := sha256.Sum256([]byte(str))
	return hex.EncodeToString(h[:])
}

// SendCode initiates registration when a password is provided, or resends the
// verification code (with cooldown) when the password is empty.
func (a *Auth) SendCode(ctx context.Context, email string, password string) (string, error) {
	const op = "auth.SendCode"

	log := a.log.With(slog.String("op", op))

	log.Info("sending verification code", slog.String("email", email))

	// Check if user already exists in DB
	_, err := a.userProvider.User(ctx, email)
	if err == nil {
		log.Info("user already exists")
		return "", fmt.Errorf("%s: %w", op, ErrUserExists)
	}
	if !errors.Is(err, storage.ErrUserNotFound) {
		log.Error("failed to check user existence", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Enforce resend cooldown if a verification is already pending.
	passHash, _, _, createdAt, err := a.verificationStorage.GetCode(ctx, email)
	if err == nil {
		if time.Since(createdAt) < resendCooldown {
			log.Info("resend cooldown not expired")
			return "", fmt.Errorf("%s: %w", op, ErrCodeCooldown)
		}
		// Cooldown passed: replace old verification data.
		if err := a.verificationStorage.DeleteCode(ctx, email); err != nil {
			log.Error("failed to delete old verification data", slog.String("error", err.Error()))
			return "", fmt.Errorf("%s: %w", op, err)
		}

		log.Info("existing verification data found, creating new code")
		code, err := a.createAndSaveCode(ctx, email, passHash)
		if err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}
		log.Info("verification code resent")

		return code, nil
	} else if !errors.Is(err, storage.ErrCodeNotFound) {
		log.Error("failed to get verification data", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Hash password
	passHash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Generate code, save verification data to Redis, and send the code to the user.
	code, err := a.createAndSaveCode(ctx, email, passHash)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("verification code sent")

	return code, nil
}

func (a *Auth) createAndSaveCode(ctx context.Context, email string, passHash []byte) (string, error) {
	const op = "auth.createAndSaveCode"

	log := a.log.With(slog.String("op", op))

	// Generate 6-digit code
	code, err := generateCode()
	if err != nil {
		log.Error("failed to generate code", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	codeHash := hashString(code)

	// Save to Redis (overwrites any existing entry for this email)
	err = a.verificationStorage.SaveCode(ctx, email, passHash, codeHash, maxAttempts, verificationTTL)
	if err != nil {
		log.Error("failed to save verification data", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return code, nil
}

// Register verifies the 6-digit code and, if valid, persists the user in the DB
// and returns an access + refresh token pair.
func (a *Auth) Register(ctx context.Context, email string, code string) (string, string, error) {
	const op = "auth.Register"

	log := a.log.With(slog.String("op", op))

	log.Info("registering user", slog.String("email", email))

	// Atomically verify the code and — on match — delete it from Redis,
	// all in one Lua transaction to prevent brute-force via concurrent requests.
	passHash, matched, err := a.verificationStorage.VerifyCode(ctx, email, hashString(code))
	if err != nil {
		if errors.Is(err, storage.ErrCodeNotFound) {
			log.Info("no pending verification found")
			return "", "", fmt.Errorf("%s: %w", op, ErrCodeNotFound)
		}
		if errors.Is(err, storage.ErrNoAttemptsLeft) {
			log.Info("no attempts left")
			return "", "", fmt.Errorf("%s: %w", op, ErrNoAttemptsLeft)
		}
		log.Error("failed to verify code", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if !matched {
		log.Info("invalid code")
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCode)
	}

	// Code is correct and already deleted from Redis by the Lua script.
	// Save user to the database.
	userID, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Info("user already exists")
			return "", "", fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save user", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

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

// SendLinkForResetPassword generates a unique link for password reset, saves it in the verification storage, and returns the link.
func (a *Auth) SendLinkForResetPassword(ctx context.Context, email string) (string, error) {
	const op = "auth.SendLinkForResetPassword"

	log := a.log.With(slog.String("op", op))

	log.Info("sending link for reset password", slog.String("email", email))

	_, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Info("user not found")
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		log.Error("failed to get user", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Generate a link for password reset
	var linkHash string
	isUniqueLink := false
	for !isUniqueLink {
		secondPartLink, err := generateLink()
		if err != nil {
			log.Error("failed to generate reset link", slog.String("error", err.Error()))
			return "", fmt.Errorf("%s: %w", op, err)
		}
		linkHash = hashString(secondPartLink)

		// Ensure the generated link is unique (not already in use)
		_, err = a.verificationStorage.GetLink(ctx, linkHash)
		if err != nil {
			if errors.Is(err, storage.ErrLinkNotFound) {
				isUniqueLink = true
				continue
			}
			log.Error("failed to check link uniqueness", slog.String("error", err.Error()))
			return "", fmt.Errorf("%s: %w", op, err)
		}
	}

	// Save the link in the verification storage
	err = a.verificationStorage.SaveLink(ctx, email, linkHash, a.linkTTL)
	if err != nil {
		log.Error("failed to save reset link", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	link := fmt.Sprintf(a.linkForResetPassword, linkHash)

	return link, nil
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
