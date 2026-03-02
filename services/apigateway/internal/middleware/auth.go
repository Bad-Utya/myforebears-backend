package middleware

import (
	"context"
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
