package customtree

import (
	"encoding/json"
	customtreepb "github.com/Bad-Utya/myforebears-backend/gen/go/customtree"
	client "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/customtree"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/grpcerr"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/response"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/middleware"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct{ c *client.Client }

func New(c *client.Client) *Handler { return &Handler{c} }
func uid(w http.ResponseWriter, r *http.Request) (int, bool) {
	x, e := middleware.UserIDFromContext(r.Context())
	if e != nil {
		response.Error(w, 401, "unauthorized", "invalid token")
		return 0, false
	}
	return x, true
}
func fail(w http.ResponseWriter, e error) {
	s, m := grpcerr.HTTPStatus(e)
	response.Error(w, s, "customtree_error", m)
}
func limit(r *http.Request) int {
	x, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if x <= 0 {
		x = 20
	}
	return x
}

type createTreeBody struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	RelationDown   string `json:"relation_down"`
	RelationUp     string `json:"relation_up"`
	RootEntityName string `json:"root_entity_name"`
}
type updateTreeBody struct {
	Name               string `json:"name"`
	Description        string `json:"description"`
	RelationDown       string `json:"relation_down"`
	RelationUp         string `json:"relation_up"`
	RootEntityID       string `json:"root_entity_id"`
	IsViewRestricted   bool   `json:"is_view_restricted"`
	IsPublicOnMainPage bool   `json:"is_public_on_main_page"`
}
type entityBody struct {
	ParentID    string `json:"parent_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
type edgeBody struct {
	ParentID string `json:"parent_id"`
	ChildID  string `json:"child_id"`
}
type emailBody struct {
	Email string `json:"email"`
}
type tagsBody struct {
	TagCodes []string `json:"tag_codes"`
}

func tagCodes(r *http.Request) []string {
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

func (h *Handler) CreateTree(w http.ResponseWriter, r *http.Request) {
	u, ok := uid(w, r)
	if !ok {
		return
	}
	var b createTreeBody
	if json.NewDecoder(r.Body).Decode(&b) != nil {
		response.Error(w, 400, "bad_request", "invalid body")
		return
	}
	x, e := h.c.CreateTree(r.Context(), &customtreepb.CreateTreeRequest{RequestUserId: int32(u), Name: b.Name, Description: b.Description, RelationDown: b.RelationDown, RelationUp: b.RelationUp, RootEntityName: b.RootEntityName})
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	u, ok := uid(w, r)
	if !ok {
		return
	}
	x, e := h.c.ListMine(r.Context(), u)
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) ListByOwner(w http.ResponseWriter, r *http.Request) {
	u, e := strconv.Atoi(chi.URLParam(r, "user_id"))
	if e != nil {
		response.Error(w, 400, "bad_request", "invalid user id")
		return
	}
	x, e := h.c.ListByOwner(r.Context(), u)
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) Random(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.Random(r.Context(), limit(r))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.Search(r.Context(), r.URL.Query().Get("q"), tagCodes(r), limit(r))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.GetTree(r.Context(), chi.URLParam(r, "tree_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, map[string]any{"tags": x.GetTree().GetTags()})
}
func (h *Handler) SetTags(w http.ResponseWriter, r *http.Request) {
	u, ok := uid(w, r)
	if !ok {
		return
	}
	var b tagsBody
	if json.NewDecoder(r.Body).Decode(&b) != nil {
		response.Error(w, 400, "bad_request", "invalid body")
		return
	}
	x, e := h.c.SetTags(r.Context(), u, chi.URLParam(r, "tree_id"), b.TagCodes)
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) GetTree(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.GetTree(r.Context(), chi.URLParam(r, "tree_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) UpdateTree(w http.ResponseWriter, r *http.Request) {
	u, ok := uid(w, r)
	if !ok {
		return
	}
	var b updateTreeBody
	if json.NewDecoder(r.Body).Decode(&b) != nil {
		response.Error(w, 400, "bad_request", "invalid body")
		return
	}
	x, e := h.c.UpdateTree(r.Context(), &customtreepb.UpdateTreeRequest{RequestUserId: int32(u), TreeId: chi.URLParam(r, "tree_id"), Name: b.Name, Description: b.Description, RelationDown: b.RelationDown, RelationUp: b.RelationUp, RootEntityId: b.RootEntityID, IsViewRestricted: b.IsViewRestricted, IsPublicOnMainPage: b.IsPublicOnMainPage})
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) DeleteTree(w http.ResponseWriter, r *http.Request) {
	u, ok := uid(w, r)
	if !ok {
		return
	}
	if e := h.c.DeleteTree(r.Context(), u, chi.URLParam(r, "tree_id")); e != nil {
		fail(w, e)
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}
func (h *Handler) Content(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.Content(r.Context(), chi.URLParam(r, "tree_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) CreateEntity(w http.ResponseWriter, r *http.Request) {
	var b entityBody
	if json.NewDecoder(r.Body).Decode(&b) != nil {
		response.Error(w, 400, "bad_request", "invalid body")
		return
	}
	x, e := h.c.CreateEntity(r.Context(), &customtreepb.CreateEntityRequest{TreeId: chi.URLParam(r, "tree_id"), ParentId: b.ParentID, Name: b.Name, Description: b.Description})
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) GetEntity(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.GetEntity(r.Context(), chi.URLParam(r, "tree_id"), chi.URLParam(r, "entity_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) ListEntities(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.ListEntities(r.Context(), chi.URLParam(r, "tree_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) UpdateEntity(w http.ResponseWriter, r *http.Request) {
	var b entityBody
	if json.NewDecoder(r.Body).Decode(&b) != nil {
		response.Error(w, 400, "bad_request", "invalid body")
		return
	}
	x, e := h.c.UpdateEntity(r.Context(), &customtreepb.UpdateEntityRequest{TreeId: chi.URLParam(r, "tree_id"), EntityId: chi.URLParam(r, "entity_id"), Name: b.Name, Description: b.Description})
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) DeleteEntity(w http.ResponseWriter, r *http.Request) {
	if e := h.c.DeleteEntity(r.Context(), chi.URLParam(r, "tree_id"), chi.URLParam(r, "entity_id")); e != nil {
		fail(w, e)
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}
func (h *Handler) AddEdge(w http.ResponseWriter, r *http.Request) {
	var b edgeBody
	if json.NewDecoder(r.Body).Decode(&b) != nil {
		response.Error(w, 400, "bad_request", "invalid body")
		return
	}
	if e := h.c.AddEdge(r.Context(), &customtreepb.AddEdgeRequest{TreeId: chi.URLParam(r, "tree_id"), ParentId: b.ParentID, ChildId: b.ChildID}); e != nil {
		fail(w, e)
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}
func (h *Handler) RemoveEdge(w http.ResponseWriter, r *http.Request) {
	var b edgeBody
	if json.NewDecoder(r.Body).Decode(&b) != nil {
		response.Error(w, 400, "bad_request", "invalid body")
		return
	}
	if e := h.c.RemoveEdge(r.Context(), &customtreepb.RemoveEdgeRequest{TreeId: chi.URLParam(r, "tree_id"), ParentId: b.ParentID, ChildId: b.ChildID}); e != nil {
		fail(w, e)
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}
func (h *Handler) AddEmail(w http.ResponseWriter, r *http.Request) {
	var b emailBody
	_ = json.NewDecoder(r.Body).Decode(&b)
	if e := h.c.AddEmail(r.Context(), chi.URLParam(r, "tree_id"), b.Email); e != nil {
		fail(w, e)
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}
func (h *Handler) ListEmails(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.ListEmails(r.Context(), chi.URLParam(r, "tree_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) DeleteEmail(w http.ResponseWriter, r *http.Request) {
	var b emailBody
	_ = json.NewDecoder(r.Body).Decode(&b)
	if e := h.c.DeleteEmail(r.Context(), chi.URLParam(r, "tree_id"), b.Email); e != nil {
		fail(w, e)
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}
func (h *Handler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	if e := r.ParseMultipartForm(16 << 20); e != nil {
		response.Error(w, 400, "bad_request", "invalid multipart")
		return
	}
	f, head, e := r.FormFile("file")
	if e != nil {
		response.Error(w, 400, "bad_request", "file required")
		return
	}
	defer f.Close()
	data, e := io.ReadAll(f)
	if e != nil {
		response.Error(w, 400, "bad_request", "read failed")
		return
	}
	x, e := h.c.UploadPhoto(r.Context(), &customtreepb.UploadPhotoRequest{TreeId: chi.URLParam(r, "tree_id"), EntityId: chi.URLParam(r, "entity_id"), FileName: head.Filename, MimeType: head.Header.Get("Content-Type"), Content: data, IsAvatar: r.FormValue("is_avatar") == "true"})
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) ListPhotos(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.ListPhotos(r.Context(), chi.URLParam(r, "tree_id"), chi.URLParam(r, "entity_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) GetPhoto(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.GetPhoto(r.Context(), chi.URLParam(r, "tree_id"), chi.URLParam(r, "entity_id"), chi.URLParam(r, "photo_id"))
	if e != nil {
		fail(w, e)
		return
	}
	w.Header().Set("Content-Type", x.GetPhoto().GetMimeType())
	_, _ = w.Write(x.GetContent())
}
func (h *Handler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	if e := h.c.DeletePhoto(r.Context(), chi.URLParam(r, "tree_id"), chi.URLParam(r, "entity_id"), chi.URLParam(r, "photo_id")); e != nil {
		fail(w, e)
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}
func (h *Handler) Coordinates(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.Coordinates(r.Context(), chi.URLParam(r, "tree_id"), r.URL.Query().Get("root_entity_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}
func (h *Handler) SVG(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.SVG(r.Context(), chi.URLParam(r, "tree_id"), r.URL.Query().Get("root_entity_id"))
	if e != nil {
		fail(w, e)
		return
	}
	w.Header().Set("Content-Type", x.GetMimeType())
	_, _ = w.Write(x.GetContent())
}
