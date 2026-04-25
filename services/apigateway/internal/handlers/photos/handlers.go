package photos

import (
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	photospb "github.com/Bad-Utya/myforebears-backend/gen/go/photos"
	photosclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/photos"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/grpcerr"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/response"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/middleware"
	"github.com/go-chi/chi/v5"
)

const maxUploadSize = 15 * 1024 * 1024

type Handler struct {
	log    *slog.Logger
	client *photosclient.Client
}

func New(log *slog.Logger, client *photosclient.Client) *Handler {
	return &Handler{log: log, client: client}
}

type photoJSON struct {
	ID             string `json:"id"`
	OwnerUserID    int32  `json:"owner_user_id"`
	TreeID         string `json:"tree_id"`
	PersonID       string `json:"person_id"`
	EventID        string `json:"event_id"`
	IsUserAvatar   bool   `json:"is_user_avatar"`
	IsPersonAvatar bool   `json:"is_person_avatar"`
	FileName       string `json:"file_name"`
	MimeType       string `json:"mime_type"`
	SizeBytes      int64  `json:"size_bytes"`
	CreatedAtUnix  int64  `json:"created_at_unix"`
}

type photoSuccessData struct {
	Photo photoJSON `json:"photo"`
}

type photosSuccessData struct {
	Photos []photoJSON `json:"photos"`
}

type photosStatusData struct {
	Status string `json:"status"`
}

type photoSuccessResponse struct {
	Data photoSuccessData `json:"data"`
}

type photosSuccessResponse struct {
	Data photosSuccessData `json:"data"`
}

type photosStatusSuccessResponse struct {
	Data photosStatusData `json:"data"`
}

// UploadUserAvatar uploads avatar image for authenticated user.
// @Summary Upload user avatar
// @Tags photos
// @Accept mpfd
// @Produce json
// @Security ApiKeyAuth
// @Param file formData file true "Avatar file"
// @Success 200 {object} photoSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/photos/user/avatar [post]
func (h *Handler) UploadUserAvatar(w http.ResponseWriter, r *http.Request) {
	userID, fileName, mimeType, content, ok := h.extractUploadPayload(w, r)
	if !ok {
		return
	}

	resp, err := h.client.UploadUserAvatar(r.Context(), &photospb.UploadUserAvatarRequest{
		RequestUserId: int32(userID),
		FileName:      fileName,
		MimeType:      mimeType,
		Content:       content,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("upload user avatar failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	response.OK(w, map[string]any{"photo": toPhotoJSON(resp.GetPhoto())})
}

// GetUserAvatar returns user avatar binary by user ID.
// @Summary Get user avatar
// @Tags photos
// @Produce octet-stream
// @Param user_id query int true "User ID"
// @Success 200 {file} binary
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/photos/user/avatar [get]
func (h *Handler) GetUserAvatar(w http.ResponseWriter, r *http.Request) {
	rawUserID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	if rawUserID == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "user_id is required")
		return
	}

	userID, err := strconv.Atoi(rawUserID)
	if err != nil || userID <= 0 {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid user_id")
		return
	}

	resp, err := h.client.GetUserAvatar(r.Context(), &photospb.GetUserAvatarRequest{RequestUserId: int32(userID)})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("get user avatar failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	h.writeBinaryPhoto(w, resp)
}

// UploadPersonAvatar uploads avatar for a person in tree.
// @Summary Upload person avatar
// @Tags photos
// @Accept mpfd
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param person_id path string true "Person ID"
// @Param file formData file true "Avatar file"
// @Success 200 {object} photoSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/photos/{tree_id}/persons/{person_id}/avatar [post]
func (h *Handler) UploadPersonAvatar(w http.ResponseWriter, r *http.Request) {
	_, fileName, mimeType, content, ok := h.extractUploadPayload(w, r)
	if !ok {
		return
	}

	treeID := chi.URLParam(r, "tree_id")
	personID := chi.URLParam(r, "person_id")
	if strings.TrimSpace(treeID) == "" || strings.TrimSpace(personID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id and person_id are required")
		return
	}

	resp, err := h.client.UploadPersonAvatar(r.Context(), &photospb.UploadPersonAvatarRequest{
		TreeId:   treeID,
		PersonId: personID,
		FileName: fileName,
		MimeType: mimeType,
		Content:  content,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("upload person avatar failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	response.OK(w, map[string]any{"photo": toPhotoJSON(resp.GetPhoto())})
}

// GetPersonAvatar returns person avatar binary.
// @Summary Get person avatar
// @Tags photos
// @Produce octet-stream
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param person_id path string true "Person ID"
// @Success 200 {file} binary
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/photos/{tree_id}/persons/{person_id}/avatar [get]
func (h *Handler) GetPersonAvatar(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	personID := chi.URLParam(r, "person_id")
	if strings.TrimSpace(treeID) == "" || strings.TrimSpace(personID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id and person_id are required")
		return
	}

	resp, err := h.client.GetPersonAvatar(r.Context(), &photospb.GetPersonAvatarRequest{
		TreeId:   treeID,
		PersonId: personID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("get person avatar failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	h.writeBinaryPhoto(w, resp)
}

// UploadPersonPhoto uploads a regular photo linked to person.
// @Summary Upload person photo
// @Tags photos
// @Accept mpfd
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param person_id path string true "Person ID"
// @Param file formData file true "Photo file"
// @Success 200 {object} photoSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/photos/{tree_id}/persons/{person_id} [post]
func (h *Handler) UploadPersonPhoto(w http.ResponseWriter, r *http.Request) {
	_, fileName, mimeType, content, ok := h.extractUploadPayload(w, r)
	if !ok {
		return
	}

	treeID := chi.URLParam(r, "tree_id")
	personID := chi.URLParam(r, "person_id")
	if strings.TrimSpace(treeID) == "" || strings.TrimSpace(personID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id and person_id are required")
		return
	}

	resp, err := h.client.UploadPersonPhoto(r.Context(), &photospb.UploadPersonPhotoRequest{
		TreeId:   treeID,
		PersonId: personID,
		FileName: fileName,
		MimeType: mimeType,
		Content:  content,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("upload person photo failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	response.OK(w, map[string]any{"photo": toPhotoJSON(resp.GetPhoto())})
}

// ListPersonPhotos lists photos linked to a person.
// @Summary List person photos
// @Tags photos
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param person_id path string true "Person ID"
// @Success 200 {object} photosSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/photos/{tree_id}/persons/{person_id} [get]
func (h *Handler) ListPersonPhotos(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	personID := chi.URLParam(r, "person_id")
	if strings.TrimSpace(treeID) == "" || strings.TrimSpace(personID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id and person_id are required")
		return
	}

	resp, err := h.client.ListPersonPhotos(r.Context(), &photospb.ListPersonPhotosRequest{
		TreeId:   treeID,
		PersonId: personID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("list person photos failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	photos := make([]map[string]any, 0, len(resp.GetPhotos()))
	for _, p := range resp.GetPhotos() {
		photos = append(photos, toPhotoJSON(p))
	}

	response.OK(w, map[string]any{"photos": photos})
}

// UploadEventPhoto uploads a photo linked to event.
// @Summary Upload event photo
// @Tags photos
// @Accept mpfd
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param event_id path string true "Event ID"
// @Param file formData file true "Photo file"
// @Success 200 {object} photoSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/photos/{tree_id}/events/{event_id} [post]
func (h *Handler) UploadEventPhoto(w http.ResponseWriter, r *http.Request) {
	_, fileName, mimeType, content, ok := h.extractUploadPayload(w, r)
	if !ok {
		return
	}

	treeID := chi.URLParam(r, "tree_id")
	eventID := chi.URLParam(r, "event_id")
	if strings.TrimSpace(treeID) == "" || strings.TrimSpace(eventID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id and event_id are required")
		return
	}

	resp, err := h.client.UploadEventPhoto(r.Context(), &photospb.UploadEventPhotoRequest{
		TreeId:   treeID,
		EventId:  eventID,
		FileName: fileName,
		MimeType: mimeType,
		Content:  content,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("upload event photo failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	response.OK(w, map[string]any{"photo": toPhotoJSON(resp.GetPhoto())})
}

// ListEventPhotos lists photos linked to event.
// @Summary List event photos
// @Tags photos
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param event_id path string true "Event ID"
// @Success 200 {object} photosSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/photos/{tree_id}/events/{event_id} [get]
func (h *Handler) ListEventPhotos(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	eventID := chi.URLParam(r, "event_id")
	if strings.TrimSpace(treeID) == "" || strings.TrimSpace(eventID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id and event_id are required")
		return
	}

	resp, err := h.client.ListEventPhotos(r.Context(), &photospb.ListEventPhotosRequest{
		TreeId:  treeID,
		EventId: eventID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("list event photos failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	photos := make([]map[string]any, 0, len(resp.GetPhotos()))
	for _, p := range resp.GetPhotos() {
		photos = append(photos, toPhotoJSON(p))
	}

	response.OK(w, map[string]any{"photos": photos})
}

// GetPhotoByID returns photo binary by photo ID.
// @Summary Get photo by ID
// @Tags photos
// @Produce octet-stream
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param photo_id path string true "Photo ID"
// @Success 200 {file} binary
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/photos/{tree_id}/{photo_id} [get]
func (h *Handler) GetPhotoByID(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	photoID := chi.URLParam(r, "photo_id")
	if strings.TrimSpace(treeID) == "" || strings.TrimSpace(photoID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id and photo_id are required")
		return
	}

	resp, err := h.client.GetPhotoByID(r.Context(), &photospb.GetPhotoByIDRequest{
		TreeId:  treeID,
		PhotoId: photoID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("get photo by id failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	h.writeBinaryPhoto(w, resp)
}

// DeletePhotoByID deletes photo by ID.
// @Summary Delete photo by ID
// @Tags photos
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param photo_id path string true "Photo ID"
// @Success 200 {object} photosStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/photos/{tree_id}/{photo_id} [delete]
func (h *Handler) DeletePhotoByID(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	photoID := chi.URLParam(r, "photo_id")
	if strings.TrimSpace(treeID) == "" || strings.TrimSpace(photoID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id and photo_id are required")
		return
	}

	err := h.client.DeletePhotoByID(r.Context(), &photospb.DeletePhotoByIDRequest{
		TreeId:  treeID,
		PhotoId: photoID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("delete photo by id failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}

func (h *Handler) extractUploadPayload(w http.ResponseWriter, r *http.Request) (int, string, string, []byte, bool) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return 0, "", "", nil, false
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid multipart form")
		return 0, "", "", nil, false
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "file field is required")
		return 0, "", "", nil, false
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, maxUploadSize+1))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "failed to read file")
		return 0, "", "", nil, false
	}
	if len(content) > maxUploadSize {
		response.Error(w, http.StatusBadRequest, "bad_request", "file too large")
		return 0, "", "", nil, false
	}

	mimeType := header.Header.Get("Content-Type")
	if strings.TrimSpace(mimeType) == "" {
		mimeType = http.DetectContentType(content)
	}

	return userID, header.Filename, mimeType, content, true
}

func (h *Handler) writeBinaryPhoto(w http.ResponseWriter, resp *photospb.GetPhotoContentResponse) {
	photo := resp.GetPhoto()
	if photo == nil {
		response.Error(w, http.StatusInternalServerError, "photos_error", "photo metadata missing")
		return
	}

	w.Header().Set("Content-Type", photo.GetMimeType())
	w.Header().Set("Content-Disposition", "inline; filename=\""+photo.GetFileName()+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp.GetContent())
}

func toPhotoJSON(p *photospb.Photo) map[string]any {
	if p == nil {
		return map[string]any{}
	}

	return map[string]any{
		"id":               p.GetId(),
		"owner_user_id":    p.GetOwnerUserId(),
		"tree_id":          p.GetTreeId(),
		"person_id":        p.GetPersonId(),
		"event_id":         p.GetEventId(),
		"is_user_avatar":   p.GetIsUserAvatar(),
		"is_person_avatar": p.GetIsPersonAvatar(),
		"file_name":        p.GetFileName(),
		"mime_type":        p.GetMimeType(),
		"size_bytes":       p.GetSizeBytes(),
		"created_at_unix":  p.GetCreatedAtUnix(),
	}
}
