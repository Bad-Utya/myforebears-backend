package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	familytreeclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/grpcerr"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/response"
	"github.com/go-chi/chi/v5"
)

type TreeAccessChecker struct {
	log          *slog.Logger
	tokenChecker *TokenChecker
	familyTree   *familytreeclient.Client
}

func NewTreeAccessChecker(log *slog.Logger, tokenChecker *TokenChecker, familyTree *familytreeclient.Client) *TreeAccessChecker {
	return &TreeAccessChecker{
		log:          log,
		tokenChecker: tokenChecker,
		familyTree:   familyTree,
	}
}

func (tc *TreeAccessChecker) ReadAccessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		treeID := strings.TrimSpace(chi.URLParam(r, "tree_id"))
		if treeID == "" {
			response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
			return
		}

		treeResp, err := tc.familyTree.GetTreeAccessInfo(r.Context(), treeID)
		if err != nil {
			tc.writeFamilyTreeError(w, err)
			return
		}
		tree := treeResp.GetTree()
		if tree == nil {
			response.Error(w, http.StatusInternalServerError, "internal_error", "tree access info is missing")
			return
		}

		if !tree.GetIsViewRestricted() {
			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), map[string]interface{}{
				"user_id": float64(tree.GetCreatorId()),
			})))
			return
		}

		claims, err := tc.tokenChecker.ClaimsFromRequest(r)
		if err != nil {
			tc.writeAuthError(w, err)
			return
		}

		requestUserID, ok := claims["user_id"].(float64)
		if !ok || requestUserID <= 0 {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
			return
		}

		email, ok := claims["email"].(string)
		if !ok || strings.TrimSpace(email) == "" {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
			return
		}

		if int32(requestUserID) == tree.GetCreatorId() {
			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
			return
		}

		allowed, err := tc.familyTree.IsTreeAccessEmailAllowed(r.Context(), treeID, email)
		if err != nil {
			tc.writeFamilyTreeError(w, err)
			return
		}
		if !allowed {
			response.Error(w, http.StatusForbidden, "forbidden", "forbidden")
			return
		}

		claims["user_id"] = float64(tree.GetCreatorId())
		next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
	})
}

func (tc *TreeAccessChecker) OwnerOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		treeID := strings.TrimSpace(chi.URLParam(r, "tree_id"))
		if treeID == "" {
			response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
			return
		}

		claims, err := tc.tokenChecker.ClaimsFromRequest(r)
		if err != nil {
			tc.writeAuthError(w, err)
			return
		}

		requestUserID, ok := claims["user_id"].(float64)
		if !ok || requestUserID <= 0 {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
			return
		}

		treeResp, err := tc.familyTree.GetTreeAccessInfo(r.Context(), treeID)
		if err != nil {
			tc.writeFamilyTreeError(w, err)
			return
		}
		tree := treeResp.GetTree()
		if tree == nil {
			response.Error(w, http.StatusInternalServerError, "internal_error", "tree access info is missing")
			return
		}

		if int32(requestUserID) != tree.GetCreatorId() {
			response.Error(w, http.StatusForbidden, "forbidden", "forbidden")
			return
		}

		next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
	})
}

func (tc *TreeAccessChecker) writeFamilyTreeError(w http.ResponseWriter, err error) {
	status, msg := grpcerr.HTTPStatus(err)
	response.Error(w, status, "familytree_error", msg)
}

func (tc *TreeAccessChecker) writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrMissingAuthorizationHeader):
		response.Error(w, http.StatusUnauthorized, "unauthorized", "missing authorization header")
	case errors.Is(err, ErrAuthInternal):
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal error")
	default:
		response.Error(w, http.StatusUnauthorized, "unauthorized", err.Error())
	}
}
