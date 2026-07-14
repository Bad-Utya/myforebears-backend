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

type customTagJSON struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
type customTreeJSON struct {
	ID                 string          `json:"id"`
	CreatorID          int32           `json:"creator_id"`
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	RelationDown       string          `json:"relation_down"`
	RelationUp         string          `json:"relation_up"`
	RootEntityID       string          `json:"root_entity_id"`
	IsViewRestricted   bool            `json:"is_view_restricted"`
	IsPublicOnMainPage bool            `json:"is_public_on_main_page"`
	CreatedAtUnix      int64           `json:"created_at_unix"`
	Tags               []customTagJSON `json:"tags"`
	SimilarityScore    float64         `json:"similarity_score"`
}
type customEntityJSON struct {
	ID            string `json:"id"`
	TreeID        string `json:"tree_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	AvatarPhotoID string `json:"avatar_photo_id"`
	CreatedAtUnix int64  `json:"created_at_unix"`
}
type customEdgeJSON struct {
	ParentID string `json:"parent_id"`
	ChildID  string `json:"child_id"`
}
type customPhotoJSON struct {
	ID            string `json:"id"`
	EntityID      string `json:"entity_id"`
	FileName      string `json:"file_name"`
	MimeType      string `json:"mime_type"`
	SizeBytes     int64  `json:"size_bytes"`
	IsAvatar      bool   `json:"is_avatar"`
	CreatedAtUnix int64  `json:"created_at_unix"`
}
type customCoordinateNodeJSON struct {
	EntityID      string  `json:"entity_id"`
	Name          string  `json:"name"`
	AvatarPhotoID string  `json:"avatar_photo_id"`
	Layer         int32   `json:"layer"`
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
}
type customCoordinateEdgeJSON struct {
	ParentID  string `json:"parent_id"`
	ChildID   string `json:"child_id"`
	LabelDown string `json:"label_down"`
	LabelUp   string `json:"label_up"`
}
type customTreeSuccessResponse struct {
	Data struct {
		Tree       customTreeJSON   `json:"tree"`
		RootEntity customEntityJSON `json:"root_entity,omitempty"`
	} `json:"data"`
}
type customTreesSuccessResponse struct {
	Data struct {
		Trees []customTreeJSON `json:"trees"`
	} `json:"data"`
}
type customTagsSuccessResponse struct {
	Data struct {
		Tags []customTagJSON `json:"tags"`
	} `json:"data"`
}
type customContentSuccessResponse struct {
	Data struct {
		Tree     customTreeJSON     `json:"tree"`
		Entities []customEntityJSON `json:"entities"`
		Edges    []customEdgeJSON   `json:"edges"`
	} `json:"data"`
}
type customEntitySuccessResponse struct {
	Data struct {
		Entity customEntityJSON `json:"entity"`
	} `json:"data"`
}
type customEntitiesSuccessResponse struct {
	Data struct {
		Entities []customEntityJSON `json:"entities"`
	} `json:"data"`
}
type customPhotoSuccessResponse struct {
	Data struct {
		Photo customPhotoJSON `json:"photo"`
	} `json:"data"`
}
type customPhotosSuccessResponse struct {
	Data struct {
		Photos []customPhotoJSON `json:"photos"`
	} `json:"data"`
}
type customEmailsSuccessResponse struct {
	Data struct {
		Emails []string `json:"emails"`
	} `json:"data"`
}
type customCoordinatesSuccessResponse struct {
	Data struct {
		Nodes  []customCoordinateNodeJSON `json:"nodes"`
		Edges  []customCoordinateEdgeJSON `json:"edges"`
		Width  float64                    `json:"width"`
		Height float64                    `json:"height"`
	} `json:"data"`
}
type customStatusSuccessResponse struct {
	Data struct {
		Status string `json:"status"`
	} `json:"data"`
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

// CreateTree creates a custom single-parent hierarchy with its root entity.
// @Summary Create custom tree
// @Tags custom-trees
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body createTreeBody true "Tree settings and root entity"
// @Success 200 {object} customTreeSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/ [post]
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

// ListMine returns all custom trees owned by the caller.
// @Summary List my custom trees
// @Tags custom-trees
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} customTreesSuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/ [get]
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

// ListByOwner returns public custom trees created by a user.
// @Summary List user's public custom trees
// @Tags custom-trees
// @Produce json
// @Param user_id path int true "Owner user ID"
// @Success 200 {object} customTreesSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/public/users/{user_id} [get]
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

// Random returns random public custom trees.
// @Summary List random public custom trees
// @Tags custom-trees
// @Produce json
// @Param limit query int false "Result count (default 20, max 100)"
// @Success 200 {object} customTreesSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/public/random [get]
func (h *Handler) Random(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.Random(r.Context(), limit(r))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}

// Search searches public custom trees by text and tag similarity.
// @Summary Search public custom trees
// @Tags custom-trees
// @Produce json
// @Param q query string false "Name or description substring"
// @Param tags query string false "Comma-separated or repeated tag codes"
// @Param limit query int false "Result count (default 20, max 100)"
// @Success 200 {object} customTreesSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/public/search [get]
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.Search(r.Context(), r.URL.Query().Get("q"), tagCodes(r), limit(r))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}

// GetTags returns tags assigned to a custom tree.
// @Summary Get custom tree tags
// @Tags custom-trees,tags
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Success 200 {object} customTagsSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/tags [get]
func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.GetTree(r.Context(), chi.URLParam(r, "tree_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, map[string]any{"tags": x.GetTree().GetTags()})
}

// SetTags replaces all tags assigned to an owned custom tree.
// @Summary Replace custom tree tags
// @Tags custom-trees,tags
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param request body tagsBody true "Complete tag-code set"
// @Success 200 {object} customTreeSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/tags [put]
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

// GetTree returns custom tree metadata.
// @Summary Get custom tree
// @Tags custom-trees
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Success 200 {object} customTreeSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id} [get]
func (h *Handler) GetTree(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.GetTree(r.Context(), chi.URLParam(r, "tree_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}

// UpdateTree updates custom tree settings and relation labels.
// @Summary Update custom tree
// @Tags custom-trees
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param request body updateTreeBody true "Complete tree settings"
// @Success 200 {object} customTreeSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id} [put]
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

// DeleteTree removes an owned custom tree and its media.
// @Summary Delete custom tree
// @Tags custom-trees
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Success 200 {object} customStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id} [delete]
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

// Content returns a custom tree with all entities and edges.
// @Summary Get custom tree content
// @Tags custom-trees
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Success 200 {object} customContentSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/content [get]
func (h *Handler) Content(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.Content(r.Context(), chi.URLParam(r, "tree_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}

// CreateEntity creates a child entity under an existing parent.
// @Summary Create custom tree entity
// @Tags custom-trees
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param request body entityBody true "Entity data and parent ID"
// @Success 200 {object} customEntitySuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/entities [post]
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

// GetEntity returns one custom tree entity.
// @Summary Get custom tree entity
// @Tags custom-trees
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param entity_id path string true "Entity ID"
// @Success 200 {object} customEntitySuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/entities/{entity_id} [get]
func (h *Handler) GetEntity(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.GetEntity(r.Context(), chi.URLParam(r, "tree_id"), chi.URLParam(r, "entity_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}

// ListEntities returns all entities in a custom tree.
// @Summary List custom tree entities
// @Tags custom-trees
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Success 200 {object} customEntitiesSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/entities [get]
func (h *Handler) ListEntities(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.ListEntities(r.Context(), chi.URLParam(r, "tree_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}

// UpdateEntity updates entity name and description.
// @Summary Update custom tree entity
// @Tags custom-trees
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param entity_id path string true "Entity ID"
// @Param request body entityBody true "Entity data"
// @Success 200 {object} customEntitySuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/entities/{entity_id} [put]
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

// DeleteEntity removes a leaf entity; roots and entities with children are protected.
// @Summary Delete custom tree entity
// @Tags custom-trees
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param entity_id path string true "Entity ID"
// @Success 200 {object} customStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/entities/{entity_id} [delete]
func (h *Handler) DeleteEntity(w http.ResponseWriter, r *http.Request) {
	if e := h.c.DeleteEntity(r.Context(), chi.URLParam(r, "tree_id"), chi.URLParam(r, "entity_id")); e != nil {
		fail(w, e)
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}

// AddEdge connects two existing entities while enforcing one parent and no cycles.
// @Summary Add custom tree edge
// @Tags custom-trees
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param request body edgeBody true "Parent and child IDs"
// @Success 200 {object} customStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/edges [post]
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

// RemoveEdge removes a directed parent-child edge.
// @Summary Remove custom tree edge
// @Tags custom-trees
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param request body edgeBody true "Parent and child IDs"
// @Success 200 {object} customStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/edges [delete]
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

// AddEmail grants read access to a restricted custom tree by email.
// @Summary Add custom tree access email
// @Tags custom-trees
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param request body emailBody true "Allowed email"
// @Success 200 {object} customStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/access-emails [post]
func (h *Handler) AddEmail(w http.ResponseWriter, r *http.Request) {
	var b emailBody
	_ = json.NewDecoder(r.Body).Decode(&b)
	if e := h.c.AddEmail(r.Context(), chi.URLParam(r, "tree_id"), b.Email); e != nil {
		fail(w, e)
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}

// ListEmails returns emails allowed to read a restricted custom tree.
// @Summary List custom tree access emails
// @Tags custom-trees
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Success 200 {object} customEmailsSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/access-emails [get]
func (h *Handler) ListEmails(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.ListEmails(r.Context(), chi.URLParam(r, "tree_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}

// DeleteEmail revokes email access to a custom tree.
// @Summary Delete custom tree access email
// @Tags custom-trees
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param request body emailBody true "Email to remove"
// @Success 200 {object} customStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/access-emails [delete]
func (h *Handler) DeleteEmail(w http.ResponseWriter, r *http.Request) {
	var b emailBody
	_ = json.NewDecoder(r.Body).Decode(&b)
	if e := h.c.DeleteEmail(r.Context(), chi.URLParam(r, "tree_id"), b.Email); e != nil {
		fail(w, e)
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}

// UploadPhoto uploads a gallery image or avatar for a custom-tree entity.
// @Summary Upload custom entity photo
// @Tags custom-trees
// @Accept mpfd
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param entity_id path string true "Entity ID"
// @Param file formData file true "Image file"
// @Param is_avatar formData bool false "Use as avatar"
// @Success 200 {object} customPhotoSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/entities/{entity_id}/photos [post]
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

// ListPhotos returns photo metadata for a custom-tree entity.
// @Summary List custom entity photos
// @Tags custom-trees
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param entity_id path string true "Entity ID"
// @Success 200 {object} customPhotosSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/entities/{entity_id}/photos [get]
func (h *Handler) ListPhotos(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.ListPhotos(r.Context(), chi.URLParam(r, "tree_id"), chi.URLParam(r, "entity_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}

// GetPhoto returns custom-tree entity photo bytes.
// @Summary Get custom entity photo
// @Tags custom-trees
// @Produce application/octet-stream
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param entity_id path string true "Entity ID"
// @Param photo_id path string true "Photo ID"
// @Success 200 {file} binary
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/entities/{entity_id}/photos/{photo_id} [get]
func (h *Handler) GetPhoto(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.GetPhoto(r.Context(), chi.URLParam(r, "tree_id"), chi.URLParam(r, "entity_id"), chi.URLParam(r, "photo_id"))
	if e != nil {
		fail(w, e)
		return
	}
	w.Header().Set("Content-Type", x.GetPhoto().GetMimeType())
	_, _ = w.Write(x.GetContent())
}

// DeletePhoto removes a photo from a custom-tree entity.
// @Summary Delete custom entity photo
// @Tags custom-trees
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param entity_id path string true "Entity ID"
// @Param photo_id path string true "Photo ID"
// @Success 200 {object} customStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/entities/{entity_id}/photos/{photo_id} [delete]
func (h *Handler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	if e := h.c.DeletePhoto(r.Context(), chi.URLParam(r, "tree_id"), chi.URLParam(r, "entity_id"), chi.URLParam(r, "photo_id")); e != nil {
		fail(w, e)
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}

// Coordinates returns BFS/annealing layout coordinates for a custom tree.
// @Summary Render custom tree coordinates
// @Tags custom-trees
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param root_entity_id query string false "Entity used as visualization root"
// @Success 200 {object} customCoordinatesSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/coordinates [get]
func (h *Handler) Coordinates(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.Coordinates(r.Context(), chi.URLParam(r, "tree_id"), r.URL.Query().Get("root_entity_id"))
	if e != nil {
		fail(w, e)
		return
	}
	response.OK(w, x)
}

// SVG renders a custom tree as an SVG document.
// @Summary Render custom tree SVG
// @Tags custom-trees
// @Produce image/svg+xml
// @Security ApiKeyAuth
// @Param tree_id path string true "Custom tree ID"
// @Param root_entity_id query string false "Entity used as visualization root"
// @Success 200 {file} binary
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/custom-trees/{tree_id}/svg [get]
func (h *Handler) SVG(w http.ResponseWriter, r *http.Request) {
	x, e := h.c.SVG(r.Context(), chi.URLParam(r, "tree_id"), r.URL.Query().Get("root_entity_id"))
	if e != nil {
		fail(w, e)
		return
	}
	w.Header().Set("Content-Type", x.GetMimeType())
	_, _ = w.Write(x.GetContent())
}
