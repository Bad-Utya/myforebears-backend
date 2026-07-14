package middleware

import (
	customclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/customtree"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/response"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
)

type CustomTreeAccess struct {
	token  *TokenChecker
	client *customclient.Client
}

func NewCustomTreeAccess(t *TokenChecker, c *customclient.Client) *CustomTreeAccess {
	return &CustomTreeAccess{t, c}
}
func (m *CustomTreeAccess) Owner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, e := m.token.ClaimsFromRequest(r)
		if e != nil {
			response.Error(w, 401, "unauthorized", e.Error())
			return
		}
		uid, ok := claims["user_id"].(float64)
		if !ok {
			response.Error(w, 401, "unauthorized", "invalid claims")
			return
		}
		t, e := m.client.GetTree(r.Context(), chi.URLParam(r, "tree_id"))
		if e != nil {
			response.Error(w, 404, "customtree_error", e.Error())
			return
		}
		if int32(uid) != t.GetTree().GetCreatorId() {
			response.Error(w, 403, "forbidden", "forbidden")
			return
		}
		next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
	})
}
func (m *CustomTreeAccess) Read(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, e := m.client.GetTree(r.Context(), chi.URLParam(r, "tree_id"))
		if e != nil {
			response.Error(w, 404, "customtree_error", e.Error())
			return
		}
		if !t.GetTree().GetIsViewRestricted() {
			next.ServeHTTP(w, r)
			return
		}
		claims, e := m.token.ClaimsFromRequest(r)
		if e != nil {
			response.Error(w, 401, "unauthorized", e.Error())
			return
		}
		uid, _ := claims["user_id"].(float64)
		if int32(uid) == t.GetTree().GetCreatorId() {
			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
			return
		}
		email, _ := claims["email"].(string)
		ok, e := m.client.EmailAllowed(r.Context(), chi.URLParam(r, "tree_id"), strings.TrimSpace(email))
		if e != nil || !ok {
			response.Error(w, 403, "forbidden", "forbidden")
			return
		}
		next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
	})
}
