package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type userStorageStub struct {
	t                *testing.T
	saveUserFn       func(ctx context.Context, email string, passHash []byte, nickname string) (int, error)
	getUserFn        func(ctx context.Context, email string) (models.User, error)
	getUserByIDFn    func(ctx context.Context, userID int) (models.User, error)
	searchUsersFn    func(ctx context.Context, query string, limit int) ([]models.User, error)
	updatePasswordFn func(ctx context.Context, email string, password []byte) error
	updateNicknameFn func(ctx context.Context, userID int, nickname string) error
	updatePrefsFn    func(ctx context.Context, userID int, language string, theme string) error
}

func (s *userStorageStub) SaveUser(ctx context.Context, email string, passHash []byte, nickname string) (int, error) {
	if s.saveUserFn == nil {
		s.t.Fatalf("unexpected SaveUser call")
	}
	return s.saveUserFn(ctx, email, passHash, nickname)
}

func (s *userStorageStub) GetUser(ctx context.Context, email string) (models.User, error) {
	if s.getUserFn == nil {
		s.t.Fatalf("unexpected GetUser call")
	}
	return s.getUserFn(ctx, email)
}

func (s *userStorageStub) GetUserByID(ctx context.Context, userID int) (models.User, error) {
	if s.getUserByIDFn == nil {
		s.t.Fatalf("unexpected GetUserByID call")
	}
	return s.getUserByIDFn(ctx, userID)
}

func (s *userStorageStub) SearchUsersByNickname(ctx context.Context, query string, limit int) ([]models.User, error) {
	if s.searchUsersFn == nil {
		s.t.Fatalf("unexpected SearchUsersByNickname call")
	}
	return s.searchUsersFn(ctx, query, limit)
}

func (s *userStorageStub) UpdatePassword(ctx context.Context, email string, password []byte) error {
	if s.updatePasswordFn == nil {
		s.t.Fatalf("unexpected UpdatePassword call")
	}
	return s.updatePasswordFn(ctx, email, password)
}

func (s *userStorageStub) UpdateNickname(ctx context.Context, userID int, nickname string) error {
	if s.updateNicknameFn == nil {
		s.t.Fatalf("unexpected UpdateNickname call")
	}
	return s.updateNicknameFn(ctx, userID, nickname)
}

func (s *userStorageStub) UpdatePreferences(ctx context.Context, userID int, language string, theme string) error {
	if s.updatePrefsFn == nil {
		s.t.Fatalf("unexpected UpdatePreferences call")
	}
	return s.updatePrefsFn(ctx, userID, language, theme)
}

type verificationStorageStub struct {
	t              *testing.T
	saveCodeFn     func(ctx context.Context, email string, passHash []byte, codeHash string, attempts int, ttl time.Duration) error
	getCodeFn      func(ctx context.Context, email string) ([]byte, string, int, time.Time, error)
	verifyCodeFn   func(ctx context.Context, email string, inputCodeHash string) ([]byte, bool, error)
	saveLinkFn     func(ctx context.Context, email string, linkHash string, ttl time.Duration) error
	getEmailByLink func(ctx context.Context, linkHash string) (string, error)
	deleteCodeFn   func(ctx context.Context, email string) error
}

func (s *verificationStorageStub) SaveCode(ctx context.Context, email string, passHash []byte, codeHash string, attempts int, ttl time.Duration) error {
	if s.saveCodeFn == nil {
		s.t.Fatalf("unexpected SaveCode call")
	}
	return s.saveCodeFn(ctx, email, passHash, codeHash, attempts, ttl)
}

func (s *verificationStorageStub) GetCode(ctx context.Context, email string) ([]byte, string, int, time.Time, error) {
	if s.getCodeFn == nil {
		s.t.Fatalf("unexpected GetCode call")
	}
	return s.getCodeFn(ctx, email)
}

func (s *verificationStorageStub) VerifyCode(ctx context.Context, email string, inputCodeHash string) ([]byte, bool, error) {
	if s.verifyCodeFn == nil {
		s.t.Fatalf("unexpected VerifyCode call")
	}
	return s.verifyCodeFn(ctx, email, inputCodeHash)
}

func (s *verificationStorageStub) SaveLink(ctx context.Context, email string, linkHash string, ttl time.Duration) error {
	if s.saveLinkFn == nil {
		s.t.Fatalf("unexpected SaveLink call")
	}
	return s.saveLinkFn(ctx, email, linkHash, ttl)
}

func (s *verificationStorageStub) GetEmailByLink(ctx context.Context, linkHash string) (string, error) {
	if s.getEmailByLink == nil {
		s.t.Fatalf("unexpected GetEmailByLink call")
	}
	return s.getEmailByLink(ctx, linkHash)
}

func (s *verificationStorageStub) DeleteCode(ctx context.Context, email string) error {
	if s.deleteCodeFn == nil {
		s.t.Fatalf("unexpected DeleteCode call")
	}
	return s.deleteCodeFn(ctx, email)
}

type accessBlacklistStub struct {
	t              *testing.T
	blacklistToken func(ctx context.Context, token string, ttl time.Duration) error
	blacklistEmail func(ctx context.Context, email string, ttl time.Duration) error
}

func (s *accessBlacklistStub) BlacklistToken(ctx context.Context, token string, ttl time.Duration) error {
	if s.blacklistToken == nil {
		s.t.Fatalf("unexpected BlacklistToken call")
	}
	return s.blacklistToken(ctx, token, ttl)
}

func (s *accessBlacklistStub) BlacklistEmail(ctx context.Context, email string, ttl time.Duration) error {
	if s.blacklistEmail == nil {
		s.t.Fatalf("unexpected BlacklistEmail call")
	}
	return s.blacklistEmail(ctx, email, ttl)
}

type emailPublisherStub struct {
	t           *testing.T
	publishCode func(ctx context.Context, to string, code string) error
	publishLink func(ctx context.Context, to string, link string) error
}

func (s *emailPublisherStub) PublishCode(ctx context.Context, to string, code string) error {
	if s.publishCode == nil {
		s.t.Fatalf("unexpected PublishCode call")
	}
	return s.publishCode(ctx, to, code)
}

func (s *emailPublisherStub) PublishLink(ctx context.Context, to string, link string) error {
	if s.publishLink == nil {
		s.t.Fatalf("unexpected PublishLink call")
	}
	return s.publishLink(ctx, to, link)
}

func TestSendCodeCreatesAndPublishesNewCode(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	var savedCodeHash string
	var savedAttempts int
	var savedTTL time.Duration
	var publishedCode string

	userStore := &userStorageStub{
		t: t,
		getUserFn: func(ctx context.Context, email string) (models.User, error) {
			return models.User{}, storage.ErrUserNotFound
		},
	}

	verificationStore := &verificationStorageStub{
		t: t,
		getCodeFn: func(ctx context.Context, email string) ([]byte, string, int, time.Time, error) {
			return nil, "", 0, time.Time{}, storage.ErrCodeNotFound
		},
		saveCodeFn: func(ctx context.Context, email string, passHash []byte, codeHash string, attempts int, ttl time.Duration) error {
			savedCodeHash = codeHash
			savedAttempts = attempts
			savedTTL = ttl
			if len(passHash) == 0 {
				t.Fatalf("expected password hash to be saved")
			}
			return nil
		},
	}

	emailPublisher := &emailPublisherStub{
		t: t,
		publishCode: func(ctx context.Context, to string, code string) error {
			publishedCode = code
			return nil
		},
	}

	a := New(log, userStore, verificationStore, "secret", time.Minute, time.Hour, "", time.Hour, &accessBlacklistStub{t: t}, emailPublisher)

	if err := a.SendCode(ctx, "user@example.com", "secret-password"); err != nil {
		t.Fatalf("SendCode error: %v", err)
	}

	if len(publishedCode) != 6 {
		t.Fatalf("expected 6-digit code, got %q", publishedCode)
	}
	if savedCodeHash != hashString(publishedCode) {
		t.Fatalf("expected code hash to match published code")
	}
	if savedAttempts != maxAttempts {
		t.Fatalf("expected attempts %d, got %d", maxAttempts, savedAttempts)
	}
	if savedTTL != verificationTTL {
		t.Fatalf("expected ttl %v, got %v", verificationTTL, savedTTL)
	}
}

func TestSendCodeRespectsCooldown(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	userStore := &userStorageStub{
		t: t,
		getUserFn: func(ctx context.Context, email string) (models.User, error) {
			return models.User{}, storage.ErrUserNotFound
		},
	}

	verificationStore := &verificationStorageStub{
		t: t,
		getCodeFn: func(ctx context.Context, email string) ([]byte, string, int, time.Time, error) {
			return []byte("hash"), "code", maxAttempts, time.Now().Add(-30 * time.Second), nil
		},
		deleteCodeFn: func(ctx context.Context, email string) error {
			t.Fatalf("unexpected DeleteCode call")
			return nil
		},
	}

	emailPublisher := &emailPublisherStub{
		t: t,
		publishCode: func(ctx context.Context, to string, code string) error {
			t.Fatalf("unexpected PublishCode call")
			return nil
		},
	}

	a := New(log, userStore, verificationStore, "secret", time.Minute, time.Hour, "", time.Hour, &accessBlacklistStub{t: t}, emailPublisher)

	err := a.SendCode(ctx, "user@example.com", "secret-password")
	if !errors.Is(err, ErrCodeCooldown) {
		t.Fatalf("expected ErrCodeCooldown, got %v", err)
	}
}

func TestRegisterCreatesUserAndTokens(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	userStore := &userStorageStub{t: t}
	userStore.saveUserFn = func(ctx context.Context, email string, passHash []byte, nickname string) (int, error) {
		if email != "user@example.com" {
			t.Fatalf("unexpected email %q", email)
		}
		if nickname != "user" {
			t.Fatalf("unexpected nickname %q", nickname)
		}
		if len(passHash) == 0 {
			t.Fatalf("expected pass hash")
		}
		return 42, nil
	}

	verificationStore := &verificationStorageStub{
		t: t,
		verifyCodeFn: func(ctx context.Context, email string, inputCodeHash string) ([]byte, bool, error) {
			if inputCodeHash != hashString("123456") {
				t.Fatalf("unexpected code hash %q", inputCodeHash)
			}
			return []byte("hash"), true, nil
		},
	}

	a := New(log, userStore, verificationStore, "secret", time.Minute, time.Hour, "", time.Hour, &accessBlacklistStub{t: t}, &emailPublisherStub{t: t})

	access, refresh, err := a.Register(ctx, "user@example.com", "123456")
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatalf("expected non-empty tokens")
	}
}

func TestLoginRejectsInvalidPassword(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	passHash, err := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt error: %v", err)
	}

	userStore := &userStorageStub{
		t: t,
		getUserFn: func(ctx context.Context, email string) (models.User, error) {
			return models.User{ID: 1, Email: email, PassHash: passHash}, nil
		},
	}

	a := New(log, userStore, &verificationStorageStub{t: t}, "secret", time.Minute, time.Hour, "", time.Hour, &accessBlacklistStub{t: t}, &emailPublisherStub{t: t})

	_, _, err = a.Login(ctx, "user@example.com", "wrong")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
