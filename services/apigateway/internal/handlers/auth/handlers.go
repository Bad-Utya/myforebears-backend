package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	authclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/auth"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/grpcerr"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/response"
)

type Handler struct {
	log    *slog.Logger
	client *authclient.Client
}

func New(log *slog.Logger, client *authclient.Client) *Handler {
	return &Handler{
		log:    log,
		client: client,
	}
}

// --- Request / Response types ---

type sendCodeRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type sendLinkForResetPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordWithLinkRequest struct {
	Link     string `json:"link"`
	Password string `json:"password"`
}

type resetPasswordWithTokenRequest struct {
	Password string `json:"password"`
}

type refreshTokensRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type tokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// --- Handlers ---

func (h *Handler) SendCode(w http.ResponseWriter, r *http.Request) {
	var req sendCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "email and password are required")
		return
	}

	code, err := h.client.SendCode(r.Context(), req.Email, req.Password)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("send code failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	response.OK(w, map[string]string{"code": code})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	if req.Email == "" || req.Code == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "email and code are required")
		return
	}

	accessToken, refreshToken, err := h.client.Register(r.Context(), req.Email, req.Code)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("register failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	response.OK(w, tokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "email and password are required")
		return
	}

	accessToken, refreshToken, err := h.client.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("login failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	response.OK(w, tokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *Handler) SendLinkForResetPassword(w http.ResponseWriter, r *http.Request) {
	var req sendLinkForResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	if req.Email == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "email is required")
		return
	}

	link, err := h.client.SendLinkForResetPassword(r.Context(), req.Email)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("send link for reset password failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	response.OK(w, map[string]string{"link": link})
}

func (h *Handler) ResetPasswordWithLink(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordWithLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	if req.Link == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "link and password are required")
		return
	}

	err := h.client.ResetPasswordWithLink(r.Context(), req.Link, req.Password)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("reset password with link failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}

func (h *Handler) ResetPasswordWithToken(w http.ResponseWriter, r *http.Request) {
	accessToken := extractBearerToken(r)
	if accessToken == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "missing or invalid authorization header")
		return
	}

	var req resetPasswordWithTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	if req.Password == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "password is required")
		return
	}

	err := h.client.ResetPasswordWithToken(r.Context(), accessToken, req.Password)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("reset password with token failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}

func (h *Handler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	var req refreshTokensRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "refresh_token is required")
		return
	}

	accessToken, refreshToken, err := h.client.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("refresh tokens failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	response.OK(w, tokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	accessToken := extractBearerToken(r)
	if accessToken == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "missing or invalid authorization header")
		return
	}

	err := h.client.Logout(r.Context(), accessToken)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("logout failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}

func (h *Handler) LogoutFromAllDevices(w http.ResponseWriter, r *http.Request) {
	accessToken := extractBearerToken(r)
	if accessToken == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "missing or invalid authorization header")
		return
	}

	err := h.client.LogoutFromAllDevices(r.Context(), accessToken)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("logout from all devices failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}

// extractBearerToken extracts the token from the Authorization header.
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
