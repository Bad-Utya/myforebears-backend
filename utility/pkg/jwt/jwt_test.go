package jwt

import (
	"testing"
	"time"
)

func TestNewTokenAndParseToken(t *testing.T) {
	secret := "super-secret"
	userID := 42
	email := "user@example.com"
	tokenType := "access"

	token, err := NewToken(userID, email, secret, time.Minute, tokenType)
	if err != nil {
		t.Fatalf("NewToken error: %v", err)
	}

	claims, err := ParseToken(token, secret)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}

	if claims["user_id"] != float64(userID) {
		t.Fatalf("expected user_id %d, got %v", userID, claims["user_id"])
	}
	if claims["email"] != email {
		t.Fatalf("expected email %q, got %v", email, claims["email"])
	}
	if claims["type"] != tokenType {
		t.Fatalf("expected type %q, got %v", tokenType, claims["type"])
	}

	createdAt, okCreated := claims["created_at"].(float64)
	exp, okExp := claims["exp"].(float64)
	if !okCreated || !okExp {
		t.Fatalf("expected numeric created_at and exp, got created_at=%v exp=%v", claims["created_at"], claims["exp"])
	}
	if exp <= createdAt {
		t.Fatalf("expected exp > created_at, got created_at=%v exp=%v", createdAt, exp)
	}
}

func TestParseTokenInvalidSecret(t *testing.T) {
	token, err := NewToken(1, "user@example.com", "secret-a", time.Minute, "access")
	if err != nil {
		t.Fatalf("NewToken error: %v", err)
	}

	if _, err := ParseToken(token, "secret-b"); err == nil {
		t.Fatalf("expected ParseToken error for invalid secret")
	}
}

func TestParseTokenInvalidToken(t *testing.T) {
	if _, err := ParseToken("not-a-token", "secret"); err == nil {
		t.Fatalf("expected ParseToken error for invalid token string")
	}
}
