package familytree

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/response"
	"github.com/go-chi/chi/v5"
)

type setTagsRequest struct {
	TagCodes []string `json:"tag_codes"`
}

type tagJSON struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type tagsSuccessResponse struct {
	Data struct {
		Tags []tagJSON `json:"tags"`
	} `json:"data"`
}

func selectedTagCodes(r *http.Request) []string {
	values := append([]string{}, r.URL.Query()["tags"]...)
	values = append(values, r.URL.Query()["tag"]...)
	result := make([]string, 0)
	for _, value := range values {
		for _, code := range strings.Split(value, ",") {
			if code = strings.TrimSpace(code); code != "" {
				result = append(result, code)
			}
		}
	}
	return result
}

// ListTags returns the shared curated tag catalog.
// @Summary List available tags
// @Tags tags
// @Produce json
// @Success 200 {object} tagsSuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/tags [get]
func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.ListTags(r.Context())
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"tags": resp.GetTags()})
}

// GetTreeTags returns tags assigned to a family tree.
// @Summary Get family tree tags
// @Tags tags,familytree
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Success 200 {object} tagsSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id}/tags [get]
func (h *Handler) GetTreeTags(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.GetTree(r.Context(), chi.URLParam(r, "tree_id"))
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"tags": resp.GetTree().GetTags()})
}

// SetTreeTags replaces all tags assigned to an owned family tree.
// @Summary Replace family tree tags
// @Tags tags,familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param request body setTagsRequest true "Complete tag-code set"
// @Success 200 {object} treeSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id}/tags [put]
func (h *Handler) SetTreeTags(w http.ResponseWriter, r *http.Request) {
	userID, ok := publicUserID(w, r)
	if !ok {
		return
	}
	var body setTagsRequest
	if json.NewDecoder(r.Body).Decode(&body) != nil {
		response.Error(w, 400, "bad_request", "invalid request body")
		return
	}
	resp, err := h.client.SetTreeTags(r.Context(), userID, chi.URLParam(r, "tree_id"), body.TagCodes)
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"tree": toTreeJSON(resp.GetTree())})
}

// GetPublicPersonTags returns tags assigned to a public person.
// @Summary Get public person tags
// @Tags tags,public-persons
// @Produce json
// @Param public_person_id path string true "Public person ID"
// @Success 200 {object} tagsSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/public-persons/{public_person_id}/tags [get]
func (h *Handler) GetPublicPersonTags(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.GetPublicPerson(r.Context(), chi.URLParam(r, "public_person_id"))
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"tags": resp.GetPerson().GetTags()})
}

// SetPublicPersonTags replaces all tags assigned to an owned public person.
// @Summary Replace public person tags
// @Tags tags,public-persons
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param public_person_id path string true "Public person ID"
// @Param request body setTagsRequest true "Complete tag-code set"
// @Success 200 {object} publicPersonSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/public-persons/{public_person_id}/tags [put]
func (h *Handler) SetPublicPersonTags(w http.ResponseWriter, r *http.Request) {
	userID, ok := publicUserID(w, r)
	if !ok {
		return
	}
	var body setTagsRequest
	if json.NewDecoder(r.Body).Decode(&body) != nil {
		response.Error(w, 400, "bad_request", "invalid request body")
		return
	}
	resp, err := h.client.SetPublicPersonTags(r.Context(), userID, chi.URLParam(r, "public_person_id"), body.TagCodes)
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"person": publicPersonJSON(resp.GetPerson())})
}
