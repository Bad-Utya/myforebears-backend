package middleware

import (
	"context"
	"net/http"
	"testing"
)

func TestExtractBearer(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	req.Header.Set("Authorization", "Bearer token-value")
	if got := extractBearer(req); got != "token-value" {
		t.Fatalf("expected token-value, got %q", got)
	}

	req.Header.Set("Authorization", "bearer token-value")
	if got := extractBearer(req); got != "token-value" {
		t.Fatalf("expected token-value case-insensitive, got %q", got)
	}

	req.Header.Set("Authorization", "Basic abc")
	if got := extractBearer(req); got != "" {
		t.Fatalf("expected empty token for non-bearer auth, got %q", got)
	}
}

func TestClaimsFromContext(t *testing.T) {
	ctx := context.Background()
	if _, err := ClaimsFromContext(ctx); err == nil {
		t.Fatalf("expected error when claims are missing")
	}

	ctx = context.WithValue(ctx, ClainsKey, "not-a-map")
	if _, err := ClaimsFromContext(ctx); err == nil {
		t.Fatalf("expected error for invalid claims type")
	}

	claims := map[string]interface{}{"user_id": float64(1)}
	ctx = context.WithValue(ctx, ClainsKey, claims)
	got, err := ClaimsFromContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["user_id"] != float64(1) {
		t.Fatalf("expected user_id in claims, got %v", got["user_id"])
	}
}

func TestUserIDFromContext(t *testing.T) {
	ctx := context.Background()
	if _, err := UserIDFromContext(ctx); err == nil {
		t.Fatalf("expected error when claims are missing")
	}

	ctx = context.WithValue(ctx, ClainsKey, map[string]interface{}{"user_id": "abc"})
	if _, err := UserIDFromContext(ctx); err == nil {
		t.Fatalf("expected error for invalid user_id type")
	}

	ctx = context.WithValue(ctx, ClainsKey, map[string]interface{}{"user_id": float64(0)})
	if _, err := UserIDFromContext(ctx); err == nil {
		t.Fatalf("expected error for non-positive user_id")
	}

	ctx = context.WithValue(ctx, ClainsKey, map[string]interface{}{"user_id": float64(10)})
	got, err := UserIDFromContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 10 {
		t.Fatalf("expected user_id 10, got %d", got)
	}
}

func TestEmailFromContext(t *testing.T) {
	ctx := context.Background()
	if _, err := EmailFromContext(ctx); err == nil {
		t.Fatalf("expected error when claims are missing")
	}

	ctx = context.WithValue(ctx, ClainsKey, map[string]interface{}{"email": 123})
	if _, err := EmailFromContext(ctx); err == nil {
		t.Fatalf("expected error for invalid email type")
	}

	ctx = context.WithValue(ctx, ClainsKey, map[string]interface{}{"email": "   "})
	if _, err := EmailFromContext(ctx); err == nil {
		t.Fatalf("expected error for empty email")
	}

	ctx = context.WithValue(ctx, ClainsKey, map[string]interface{}{"email": " User@Example.com  "})
	got, err := EmailFromContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "User@Example.com" {
		t.Fatalf("expected trimmed email, got %q", got)
	}
}
