package storage

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUserExists     = errors.New("user already exists")
	ErrCodeNotFound   = errors.New("verification code not found")
	ErrLinkNotFound   = errors.New("verification link not found")
	ErrNoAttemptsLeft = errors.New("no attempts left")
)
