package photos

import (
	"io"
	"log/slog"
	"net/http"
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

func (h *Handler) GetUserAvatar(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
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

func (h *Handler) UploadPersonAvatar(w http.ResponseWriter, r *http.Request) {
	userID, fileName, mimeType, content, ok := h.extractUploadPayload(w, r)
	if !ok {
		return
	}

	personID := chi.URLParam(r, "person_id")
	if strings.TrimSpace(personID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id and person_id are required")
		return
	}

	resp, err := h.client.UploadPersonAvatar(r.Context(), &photospb.UploadPersonAvatarRequest{
		RequestUserId: int32(userID),
		PersonId:      personID,
		FileName:      fileName,
		MimeType:      mimeType,
		Content:       content,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("upload person avatar failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	response.OK(w, map[string]any{"photo": toPhotoJSON(resp.GetPhoto())})
}

func (h *Handler) GetPersonAvatar(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	personID := chi.URLParam(r, "person_id")
	if strings.TrimSpace(personID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id and person_id are required")
		return
	}

	resp, err := h.client.GetPersonAvatar(r.Context(), &photospb.GetPersonAvatarRequest{
		RequestUserId: int32(userID),
		PersonId:      personID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("get person avatar failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	h.writeBinaryPhoto(w, resp)
}

func (h *Handler) UploadPersonPhoto(w http.ResponseWriter, r *http.Request) {
	userID, fileName, mimeType, content, ok := h.extractUploadPayload(w, r)
	if !ok {
		return
	}

	personID := chi.URLParam(r, "person_id")
	if strings.TrimSpace(personID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id and person_id are required")
		return
	}

	resp, err := h.client.UploadPersonPhoto(r.Context(), &photospb.UploadPersonPhotoRequest{
		RequestUserId: int32(userID),
		PersonId:      personID,
		FileName:      fileName,
		MimeType:      mimeType,
		Content:       content,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("upload person photo failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	response.OK(w, map[string]any{"photo": toPhotoJSON(resp.GetPhoto())})
}

func (h *Handler) ListPersonPhotos(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	personID := chi.URLParam(r, "person_id")
	if strings.TrimSpace(personID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id and person_id are required")
		return
	}

	resp, err := h.client.ListPersonPhotos(r.Context(), &photospb.ListPersonPhotosRequest{
		RequestUserId: int32(userID),
		PersonId:      personID,
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

func (h *Handler) UploadEventPhoto(w http.ResponseWriter, r *http.Request) {
	userID, fileName, mimeType, content, ok := h.extractUploadPayload(w, r)
	if !ok {
		return
	}

	eventID := chi.URLParam(r, "event_id")
	if strings.TrimSpace(eventID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "event_id is required")
		return
	}

	resp, err := h.client.UploadEventPhoto(r.Context(), &photospb.UploadEventPhotoRequest{
		RequestUserId: int32(userID),
		EventId:       eventID,
		FileName:      fileName,
		MimeType:      mimeType,
		Content:       content,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("upload event photo failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	response.OK(w, map[string]any{"photo": toPhotoJSON(resp.GetPhoto())})
}

func (h *Handler) ListEventPhotos(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	eventID := chi.URLParam(r, "event_id")
	if strings.TrimSpace(eventID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "event_id is required")
		return
	}

	resp, err := h.client.ListEventPhotos(r.Context(), &photospb.ListEventPhotosRequest{
		RequestUserId: int32(userID),
		EventId:       eventID,
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

func (h *Handler) GetPhotoByID(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	photoID := chi.URLParam(r, "photo_id")
	if strings.TrimSpace(photoID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "photo_id is required")
		return
	}

	resp, err := h.client.GetPhotoByID(r.Context(), &photospb.GetPhotoByIDRequest{
		RequestUserId: int32(userID),
		PhotoId:       photoID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("get photo by id failed", slog.String("error", err.Error()))
		response.Error(w, status, "photos_error", msg)
		return
	}

	h.writeBinaryPhoto(w, resp)
}

func (h *Handler) DeletePhotoByID(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	photoID := chi.URLParam(r, "photo_id")
	if strings.TrimSpace(photoID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "photo_id is required")
		return
	}

	err = h.client.DeletePhotoByID(r.Context(), &photospb.DeletePhotoByIDRequest{
		RequestUserId: int32(userID),
		PhotoId:       photoID,
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
