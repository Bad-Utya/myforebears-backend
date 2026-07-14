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

func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.ListTags(r.Context())
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"tags": resp.GetTags()})
}

func (h *Handler) GetTreeTags(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.GetTree(r.Context(), chi.URLParam(r, "tree_id"))
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"tags": resp.GetTree().GetTags()})
}

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

func (h *Handler) GetPublicPersonTags(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.GetPublicPerson(r.Context(), chi.URLParam(r, "public_person_id"))
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"tags": resp.GetPerson().GetTags()})
}

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
