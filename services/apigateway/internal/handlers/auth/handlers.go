package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	authclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/auth"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/grpcerr"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/response"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/middleware"
	"github.com/go-chi/chi/v5"
)

const (
	refreshTokenCookie = "refresh_token"
	refreshTokenTTL    = 30 * 24 * time.Hour
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

type updateNicknameRequest struct {
	Nickname string `json:"nickname"`
}

type tokensResponse struct {
	AccessToken string `json:"access_token"`
}

type userInfoResponse struct {
	ID       int32  `json:"id"`
	Nickname string `json:"nickname"`
}

type authStatusData struct {
	Status string `json:"status"`
}

type authStatusSuccessResponse struct {
	Data authStatusData `json:"data"`
}

type authTokensSuccessResponse struct {
	Data tokensResponse `json:"data"`
}

// --- Handlers ---

// SendCode sends a verification code to email for registration flow.
// @Summary Send verification code
// @Description Sends a verification code to the provided email.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body sendCodeRequest true "Request body"
// @Success 200 {object} authStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/send-code [post]
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

	err := h.client.SendCode(r.Context(), req.Email, req.Password)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("send code failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}

// Register registers a user by email and verification code.
// @Summary Register user
// @Description Completes registration and returns access token plus refresh cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body registerRequest true "Request body"
// @Success 200 {object} authTokensSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/register [post]
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

	setRefreshTokenCookie(w, refreshToken)
	response.OK(w, tokensResponse{AccessToken: accessToken})
}

// Login authenticates a user with email and password.
// @Summary Login user
// @Description Authenticates user credentials and returns access token plus refresh cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "Request body"
// @Success 200 {object} authTokensSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/login [post]
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

	setRefreshTokenCookie(w, refreshToken)
	response.OK(w, tokensResponse{AccessToken: accessToken})
}

// SendLinkForResetPassword sends password reset link to email.
// @Summary Send reset link
// @Description Sends a password reset link to the specified email.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body sendLinkForResetPasswordRequest true "Request body"
// @Success 200 {object} authStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/send-link-for-reset-password [post]
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

	err := h.client.SendLinkForResetPassword(r.Context(), req.Email)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("send link for reset password failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}

// ResetPasswordWithLink resets password using reset link.
// @Summary Reset password by link
// @Description Resets password with one-time link received by email.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body resetPasswordWithLinkRequest true "Request body"
// @Success 200 {object} authStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/reset-password-with-link [post]
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

// ResetPasswordWithToken resets password using current access token.
// @Summary Reset password by token
// @Description Resets password for authenticated user.
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body resetPasswordWithTokenRequest true "Request body"
// @Success 200 {object} authStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/reset-password-with-token [post]
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

// RefreshTokens refreshes access and refresh tokens by refresh cookie.
// @Summary Refresh tokens
// @Description Uses refresh cookie and returns a new access token plus rotated refresh cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} authTokensSuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/refresh [post]
func (h *Handler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshTokenCookie)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "refresh_token cookie is missing")
		return
	}

	refreshToken := strings.TrimSpace(cookie.Value)
	if refreshToken == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "refresh_token cookie is empty")
		return
	}

	accessToken, newRefreshToken, err := h.client.RefreshTokens(r.Context(), refreshToken)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("refresh tokens failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	setRefreshTokenCookie(w, newRefreshToken)
	response.OK(w, tokensResponse{AccessToken: accessToken})
}

// Logout revokes current session tokens.
// @Summary Logout user
// @Description Logs out current user and clears refresh cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} authStatusSuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/logout [post]
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

	clearRefreshTokenCookie(w)
	response.OK(w, map[string]string{"status": "ok"})
}

// LogoutFromAllDevices revokes all user sessions.
// @Summary Logout from all devices
// @Description Logs out user from all active devices and clears refresh cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} authStatusSuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/logout-all [post]
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

	clearRefreshTokenCookie(w)
	response.OK(w, map[string]string{"status": "ok"})
}

// GetUserInfo returns public user info by id.
// @Summary Get user info
// @Tags users
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {object} map[string]userInfoResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/users/{user_id} [get]
func (h *Handler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	userIDRaw := strings.TrimSpace(chi.URLParam(r, "user_id"))
	if userIDRaw == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "user_id is required")
		return
	}

	userID, err := strconv.Atoi(userIDRaw)
	if err != nil || userID <= 0 {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid user_id")
		return
	}

	resp, err := h.client.GetUserInfo(r.Context(), userID)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("get user info failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	response.OK(w, map[string]any{"user": map[string]any{
		"id":       resp.GetUser().GetId(),
		"nickname": resp.GetUser().GetNickname(),
	}})
}

// UpdateNickname updates nickname for authenticated user.
// @Summary Update my nickname
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body updateNicknameRequest true "Request body"
// @Success 200 {object} map[string]userInfoResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/users/me/nickname [patch]
func (h *Handler) UpdateNickname(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	var req updateNicknameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	resp, err := h.client.UpdateNickname(r.Context(), userID, req.Nickname)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("update nickname failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	response.OK(w, map[string]any{"user": map[string]any{
		"id":       resp.GetUser().GetId(),
		"nickname": resp.GetUser().GetNickname(),
	}})
}

// GetMe returns the authenticated user's info.
// @Summary Get user info
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]userInfoResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/users/me [get]
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	resp, err := h.client.GetUserInfo(r.Context(), userID)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("get user info failed", slog.String("error", err.Error()))
		response.Error(w, status, "auth_error", msg)
		return
	}

	response.OK(w, map[string]any{"user": map[string]any{
		"id":       resp.GetUser().GetId(),
		"nickname": resp.GetUser().GetNickname(),
	}})
}

// setRefreshTokenCookie sets the refresh token as an HttpOnly cookie.
func setRefreshTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int(refreshTokenTTL.Seconds()),
		HttpOnly: true,
		Secure:   true, // set to true in production (HTTPS)
		SameSite: http.SameSiteNoneMode,
	})
}

// clearRefreshTokenCookie removes the refresh token cookie.
func clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookie,
		Value:    "",
		Path:     "/api/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
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
