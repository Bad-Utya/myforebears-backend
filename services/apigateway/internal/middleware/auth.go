package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	redisclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/redis"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/response"
	"github.com/Bad-Utya/myforebears-backend/utility/pkg/jwt"
)

type contextKey string

const ClainsKey contextKey = "claims"

type TokenChecker struct {
	redis     *redisclient.Client
	jwtSecret string
	log       *slog.Logger
}

func NewTokenChecker(redis *redisclient.Client, jwtSecret string, log *slog.Logger) *TokenChecker {
	return &TokenChecker{
		redis:     redis,
		jwtSecret: jwtSecret,
		log:       log,
	}
}

func (tc *TokenChecker) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearer(r)
		if token == "" {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "missing authorization header")
			return
		}

		claims, err := jwt.ParseToken(token, tc.jwtSecret)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token")
			return
		}

		email, ok := claims["email"].(string)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
			return
		}

		createdAt, _ := claims["created_at"].(float64)

		isBlacklisted, err := tc.redis.IsTokenBlacklisted(r.Context(), token, email, int64(createdAt))
		if err != nil {
			tc.log.Error("failed to check token blacklist", slog.String("error", err.Error()))
			response.Error(w, http.StatusInternalServerError, "internal_error", "internal error")
			return
		}
		if isBlacklisted {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "token has been revoked")
			return
		}

		ctx := context.WithValue(r.Context(), ClainsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func ClaimsFromContext(ctx context.Context) (map[string]interface{}, error) {
	v := ctx.Value(ClainsKey)
	if v == nil {
		return nil, fmt.Errorf("claims are missing in context")
	}

	claims, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("claims have invalid type")
	}

	return claims, nil
}

func UserIDFromContext(ctx context.Context) (int, error) {
	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		return 0, err
	}

	// JWT numeric claims are decoded as float64.
	userIDRaw, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user_id claim is missing or invalid")
	}

	if userIDRaw <= 0 {
		return 0, fmt.Errorf("user_id claim must be positive")
	}

	return int(userIDRaw), nil
}
