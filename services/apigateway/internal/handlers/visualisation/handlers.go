package visualisation

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	visualisationpb "github.com/Bad-Utya/myforebears-backend/gen/go/visualisation"
	visualisationclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/visualisation"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/grpcerr"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	log    *slog.Logger
	client *visualisationclient.Client
}

func New(log *slog.Logger, client *visualisationclient.Client) *Handler {
	return &Handler{log: log, client: client}
}

type createVisualisationRequest struct {
	RootPersonID      string   `json:"root_person_id"`
	IncludedPersonIDs []string `json:"included_person_ids,omitempty"`
}

type renderCoordinatesForClientRequest struct {
	RootPersonID string `json:"root_person_id"`
	MaxDepth     int32  `json:"max_depth"`
}

// CreateAncestorsVisualisation enqueues an ancestors visualisation.
// @Summary Create ancestors visualisation
// @Tags visualisations
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param request body createVisualisationRequest true "Create visualisation request"
// @Success 202 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/visualisations/{tree_id}/ancestors [post]
func (h *Handler) CreateAncestorsVisualisation(w http.ResponseWriter, r *http.Request) {
	treeID, req, ok := h.extractCreateRequest(w, r)
	if !ok {
		return
	}

	resp, err := h.client.CreateAncestorsVisualisation(r.Context(), &visualisationpb.CreateLineageVisualisationRequest{
		TreeId:       treeID,
		RootPersonId: req.RootPersonID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("create ancestors visualisation failed", slog.String("error", err.Error()))
		response.Error(w, status, "visualisation_error", msg)
		return
	}

	response.JSON(w, http.StatusAccepted, response.SuccessResponse{Data: map[string]any{"status": resp.GetStatus(), "visualisation": toVisualisationJSON(resp.GetVisualisation())}})
}

// CreateDescendantsVisualisation enqueues a descendants visualisation.
// @Summary Create descendants visualisation
// @Tags visualisations
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param request body createVisualisationRequest true "Create visualisation request"
// @Success 202 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/visualisations/{tree_id}/descendants [post]
func (h *Handler) CreateDescendantsVisualisation(w http.ResponseWriter, r *http.Request) {
	treeID, req, ok := h.extractCreateRequest(w, r)
	if !ok {
		return
	}

	resp, err := h.client.CreateDescendantsVisualisation(r.Context(), &visualisationpb.CreateLineageVisualisationRequest{
		TreeId:       treeID,
		RootPersonId: req.RootPersonID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("create descendants visualisation failed", slog.String("error", err.Error()))
		response.Error(w, status, "visualisation_error", msg)
		return
	}

	response.JSON(w, http.StatusAccepted, response.SuccessResponse{Data: map[string]any{"status": resp.GetStatus(), "visualisation": toVisualisationJSON(resp.GetVisualisation())}})
}

// CreateAncestorsAndDescendantsVisualisation enqueues a combined visualisation.
// @Summary Create ancestors and descendants visualisation
// @Tags visualisations
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param request body createVisualisationRequest true "Create visualisation request"
// @Success 202 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/visualisations/{tree_id}/ancestors-descendants [post]
func (h *Handler) CreateAncestorsAndDescendantsVisualisation(w http.ResponseWriter, r *http.Request) {
	treeID, req, ok := h.extractCreateRequest(w, r)
	if !ok {
		return
	}

	resp, err := h.client.CreateAncestorsAndDescendantsVisualisation(r.Context(), &visualisationpb.CreateLineageVisualisationRequest{
		TreeId:       treeID,
		RootPersonId: req.RootPersonID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("create ancestors and descendants visualisation failed", slog.String("error", err.Error()))
		response.Error(w, status, "visualisation_error", msg)
		return
	}

	response.JSON(w, http.StatusAccepted, response.SuccessResponse{Data: map[string]any{"status": resp.GetStatus(), "visualisation": toVisualisationJSON(resp.GetVisualisation())}})
}

// CreateFullVisualisation enqueues a full visualisation with selected persons.
// @Summary Create full visualisation
// @Tags visualisations
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param request body createVisualisationRequest true "Create visualisation request"
// @Success 202 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/visualisations/{tree_id}/full [post]
func (h *Handler) CreateFullVisualisation(w http.ResponseWriter, r *http.Request) {
	treeID, req, ok := h.extractCreateRequest(w, r)
	if !ok {
		return
	}

	resp, err := h.client.CreateFullVisualisation(r.Context(), &visualisationpb.CreateFullVisualisationRequest{
		TreeId:            treeID,
		RootPersonId:      req.RootPersonID,
		IncludedPersonIds: req.IncludedPersonIDs,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("create full visualisation failed", slog.String("error", err.Error()))
		response.Error(w, status, "visualisation_error", msg)
		return
	}

	response.JSON(w, http.StatusAccepted, response.SuccessResponse{Data: map[string]any{"status": resp.GetStatus(), "visualisation": toVisualisationJSON(resp.GetVisualisation())}})
}

// ListTreeVisualisations returns visualisations for a tree.
// @Summary List tree visualisations
// @Tags visualisations
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/visualisations/{tree_id} [get]
func (h *Handler) ListTreeVisualisations(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	if strings.TrimSpace(treeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	resp, err := h.client.ListTreeVisualisations(r.Context(), &visualisationpb.ListTreeVisualisationsRequest{TreeId: treeID})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("list visualisations failed", slog.String("error", err.Error()))
		response.Error(w, status, "visualisation_error", msg)
		return
	}

	items := make([]map[string]any, 0, len(resp.GetVisualisations()))
	for _, item := range resp.GetVisualisations() {
		items = append(items, toVisualisationJSON(item))
	}

	response.OK(w, map[string]any{"visualisations": items})
}

// GetVisualisationByID downloads a ready visualisation PDF by ID.
// @Summary Get visualisation by ID
// @Tags visualisations
// @Produce octet-stream
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param visualisation_id path string true "Visualisation ID"
// @Success 200 {file} binary
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/visualisations/{tree_id}/{visualisation_id} [get]
func (h *Handler) GetVisualisationByID(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	visualisationID := chi.URLParam(r, "visualisation_id")
	if strings.TrimSpace(treeID) == "" || strings.TrimSpace(visualisationID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id and visualisation_id are required")
		return
	}

	resp, err := h.client.GetVisualisationByID(r.Context(), &visualisationpb.GetVisualisationByIDRequest{
		TreeId:          treeID,
		VisualisationId: visualisationID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("get visualisation by id failed", slog.String("error", err.Error()))
		response.Error(w, status, "visualisation_error", msg)
		return
	}

	h.writeBinaryVisualisation(w, resp)
}

// RenderCoordinatesForClient renders family tree as JSON coordinates for client-side rendering.
// @Summary Render coordinates for client
// @Tags visualisations
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param request body renderCoordinatesForClientRequest true "Render coordinates request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/visualisations/{tree_id}/coordinates [post]
func (h *Handler) RenderCoordinatesForClient(w http.ResponseWriter, r *http.Request) {
	treeID := strings.TrimSpace(chi.URLParam(r, "tree_id"))
	if treeID == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	var req renderCoordinatesForClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid json body")
		return
	}

	if strings.TrimSpace(req.RootPersonID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "root_person_id is required")
		return
	}
	if req.MaxDepth < 0 {
		response.Error(w, http.StatusBadRequest, "bad_request", "max_depth must be >= 0")
		return
	}

	resp, err := h.client.RenderCoordinatesForClient(r.Context(), &visualisationpb.RenderCoordinatesForClientRequest{
		TreeId:       treeID,
		RootPersonId: req.RootPersonID,
		MaxDepth:     req.MaxDepth,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("render coordinates for client failed", slog.String("error", err.Error()))
		response.Error(w, status, "visualisation_error", msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp.GetCoordinatesJson())
}

// DeleteVisualisationByID deletes visualisation by ID.
// @Summary Delete visualisation by ID
// @Tags visualisations
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param visualisation_id path string true "Visualisation ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/visualisations/{tree_id}/{visualisation_id} [delete]
func (h *Handler) DeleteVisualisationByID(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	visualisationID := chi.URLParam(r, "visualisation_id")
	if strings.TrimSpace(treeID) == "" || strings.TrimSpace(visualisationID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id and visualisation_id are required")
		return
	}

	err := h.client.DeleteVisualisationByID(r.Context(), &visualisationpb.DeleteVisualisationByIDRequest{
		TreeId:          treeID,
		VisualisationId: visualisationID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("delete visualisation by id failed", slog.String("error", err.Error()))
		response.Error(w, status, "visualisation_error", msg)
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}

func (h *Handler) extractCreateRequest(w http.ResponseWriter, r *http.Request) (string, createVisualisationRequest, bool) {
	treeID := strings.TrimSpace(chi.URLParam(r, "tree_id"))
	if treeID == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return "", createVisualisationRequest{}, false
	}

	var req createVisualisationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid json body")
		return "", createVisualisationRequest{}, false
	}

	if strings.TrimSpace(req.RootPersonID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "root_person_id is required")
		return "", createVisualisationRequest{}, false
	}

	return treeID, req, true
}

func (h *Handler) writeBinaryVisualisation(w http.ResponseWriter, resp *visualisationpb.GetVisualisationContentResponse) {
	vis := resp.GetVisualisation()
	if vis == nil {
		response.Error(w, http.StatusInternalServerError, "visualisation_error", "visualisation metadata missing")
		return
	}

	w.Header().Set("Content-Type", vis.GetMimeType())
	w.Header().Set("Content-Disposition", "inline; filename=\""+vis.GetFileName()+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp.GetContent())
}

func toVisualisationJSON(v *visualisationpb.Visualisation) map[string]any {
	if v == nil {
		return map[string]any{}
	}

	return map[string]any{
		"id":                  v.GetId(),
		"owner_user_id":       v.GetOwnerUserId(),
		"tree_id":             v.GetTreeId(),
		"root_person_id":      v.GetRootPersonId(),
		"included_person_ids": v.GetIncludedPersonIds(),
		"type":                v.GetType().String(),
		"status":              v.GetStatus().String(),
		"file_name":           v.GetFileName(),
		"mime_type":           v.GetMimeType(),
		"size_bytes":          v.GetSizeBytes(),
		"error_message":       v.GetErrorMessage(),
		"created_at_unix":     v.GetCreatedAtUnix(),
		"updated_at_unix":     v.GetUpdatedAtUnix(),
		"completed_at_unix":   v.GetCompletedAtUnix(),
	}
}
