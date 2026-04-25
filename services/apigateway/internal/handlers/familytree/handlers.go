package familytree

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	familytreeclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/grpcerr"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/response"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	log    *slog.Logger
	client *familytreeclient.Client
}

func New(log *slog.Logger, client *familytreeclient.Client) *Handler {
	return &Handler{log: log, client: client}
}

type addParentRequest struct {
	ChildID    string `json:"child_id"`
	Role       string `json:"role" enums:"FATHER,MOTHER"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Patronymic string `json:"patronymic"`
}

type addChildRequest struct {
	Parent1ID  string `json:"parent1_id"`
	Parent2ID  string `json:"parent2_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Patronymic string `json:"patronymic"`
	Gender     string `json:"gender" enums:"MALE,FEMALE"`
}

type addPartnerRequest struct {
	PersonID   string `json:"person_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Patronymic string `json:"patronymic"`
}

type updatePersonNameRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Patronymic string `json:"patronymic"`
}

type updateTreeSettingsRequest struct {
	IsViewRestricted   bool `json:"is_view_restricted"`
	IsPublicOnMainPage bool `json:"is_public_on_main_page"`
}

type treeAccessEmailRequest struct {
	Email string `json:"email"`
}

type familyTreeJSON struct {
	ID                 string `json:"id"`
	CreatorID          int32  `json:"creator_id"`
	CreatedAtUnix      int64  `json:"created_at_unix"`
	IsViewRestricted   bool   `json:"is_view_restricted"`
	IsPublicOnMainPage bool   `json:"is_public_on_main_page"`
}

type familyPersonJSON struct {
	ID            string `json:"id"`
	TreeID        string `json:"tree_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Patronymic    string `json:"patronymic"`
	Gender        string `json:"gender" enums:"GENDER_UNSPECIFIED,GENDER_MALE,GENDER_FEMALE"`
	AvatarPhotoID string `json:"avatar_photo_id"`
}

type familyRelationshipJSON struct {
	PersonIDFrom string `json:"person_id_from"`
	PersonIDTo   string `json:"person_id_to"`
	Type         string `json:"type" enums:"RELATIONSHIP_TYPE_UNSPECIFIED,RELATIONSHIP_PARENT_CHILD,RELATIONSHIP_PARTNER,RELATIONSHIP_PARTNER_UNMARRIED,RELATIONSHIP_PARTNER_MARRIED,RELATIONSHIP_PARTNER_DIVORCED"`
}

type familyStatusData struct {
	Status string `json:"status"`
}

type createTreeSuccessData struct {
	Tree       familyTreeJSON   `json:"tree"`
	RootPerson familyPersonJSON `json:"root_person"`
}

type treeSuccessData struct {
	Tree familyTreeJSON `json:"tree"`
}

type treesSuccessData struct {
	Trees []familyTreeJSON `json:"trees"`
}

type treeContentSuccessData struct {
	Persons       []familyPersonJSON       `json:"persons"`
	Relationships []familyRelationshipJSON `json:"relationships"`
}

type treeAccessEmailsSuccessData struct {
	Emails []string `json:"emails"`
}

type personsSuccessData struct {
	Persons []familyPersonJSON `json:"persons"`
}

type personSuccessData struct {
	Person familyPersonJSON `json:"person"`
}

type addParentSuccessData struct {
	Parent                  familyPersonJSON  `json:"parent"`
	AutoCreatedSecondParent *familyPersonJSON `json:"auto_created_second_parent,omitempty"`
}

type addChildSuccessData struct {
	Child             familyPersonJSON  `json:"child"`
	AutoCreatedParent *familyPersonJSON `json:"auto_created_parent,omitempty"`
}

type addPartnerSuccessData struct {
	Partner familyPersonJSON `json:"partner"`
}

type familyStatusSuccessResponse struct {
	Data familyStatusData `json:"data"`
}

type createTreeSuccessResponse struct {
	Data createTreeSuccessData `json:"data"`
}

type treeSuccessResponse struct {
	Data treeSuccessData `json:"data"`
}

type treesSuccessResponse struct {
	Data treesSuccessData `json:"data"`
}

type treeContentSuccessResponse struct {
	Data treeContentSuccessData `json:"data"`
}

type treeAccessEmailsSuccessResponse struct {
	Data treeAccessEmailsSuccessData `json:"data"`
}

type personsSuccessResponse struct {
	Data personsSuccessData `json:"data"`
}

type personSuccessResponse struct {
	Data personSuccessData `json:"data"`
}

type addParentSuccessResponse struct {
	Data addParentSuccessData `json:"data"`
}

type addChildSuccessResponse struct {
	Data addChildSuccessData `json:"data"`
}

type addPartnerSuccessResponse struct {
	Data addPartnerSuccessData `json:"data"`
}

// CreateTree creates a new family tree with root person.
// @Summary Create tree
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} createTreeSuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/ [post]
func (h *Handler) CreateTree(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	resp, err := h.client.CreateTree(r.Context(), userID)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("create tree failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	response.OK(w, map[string]any{
		"tree":        toTreeJSON(resp.GetTree()),
		"root_person": toPersonJSON(resp.GetRootPerson()),
	})
}

// ListTrees returns trees created by authenticated user.
// @Summary List trees
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} treesSuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/ [get]
func (h *Handler) ListTrees(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	resp, err := h.client.ListTreesByCreator(r.Context(), userID)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("list trees failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	trees := make([]map[string]any, 0, len(resp.GetTrees()))
	for _, t := range resp.GetTrees() {
		trees = append(trees, toTreeJSON(t))
	}

	response.OK(w, map[string]any{"trees": trees})
}

// GetTree returns a tree by ID.
// @Summary Get tree
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Success 200 {object} treeSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id} [get]
func (h *Handler) GetTree(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	if strings.TrimSpace(treeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	resp, err := h.client.GetTree(r.Context(), treeID)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("get tree failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	response.OK(w, map[string]any{"tree": toTreeJSON(resp.GetTree())})
}

// GetTreeContent returns persons and relationships of a tree.
// @Summary Get tree content
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Success 200 {object} treeContentSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id}/content [get]
func (h *Handler) GetTreeContent(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	if strings.TrimSpace(treeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	resp, err := h.client.GetTreeContent(r.Context(), treeID)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("get tree content failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	persons := make([]map[string]any, 0, len(resp.GetPersons()))
	for _, p := range resp.GetPersons() {
		persons = append(persons, toPersonJSON(p))
	}

	relationships := make([]map[string]any, 0, len(resp.GetRelationships()))
	for _, rel := range resp.GetRelationships() {
		relationships = append(relationships, map[string]any{
			"person_id_from": rel.GetPersonIdFrom(),
			"person_id_to":   rel.GetPersonIdTo(),
			"type":           rel.GetType().String(),
		})
	}

	response.OK(w, map[string]any{"persons": persons, "relationships": relationships})
}

// UpdateTreeSettings updates visibility settings of a tree.
// @Summary Update tree settings
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param request body updateTreeSettingsRequest true "Request body"
// @Success 200 {object} treeSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id} [patch]
func (h *Handler) UpdateTreeSettings(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	if strings.TrimSpace(treeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	var req updateTreeSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	resp, err := h.client.UpdateTreeSettings(r.Context(), treeID, req.IsViewRestricted, req.IsPublicOnMainPage)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("update tree settings failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	response.OK(w, map[string]any{"tree": toTreeJSON(resp.GetTree())})
}

// AddTreeAccessEmail grants read access to email for a tree.
// @Summary Add tree access email
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param request body treeAccessEmailRequest true "Request body"
// @Success 200 {object} familyStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id}/access-emails [post]
func (h *Handler) AddTreeAccessEmail(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	if strings.TrimSpace(treeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	var req treeAccessEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	err := h.client.AddTreeAccessEmail(r.Context(), treeID, req.Email)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("add tree access email failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}

// ListTreeAccessEmails lists emails with granted tree access.
// @Summary List tree access emails
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Success 200 {object} treeAccessEmailsSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id}/access-emails [get]
func (h *Handler) ListTreeAccessEmails(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	if strings.TrimSpace(treeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	resp, err := h.client.ListTreeAccessEmails(r.Context(), treeID)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("list tree access emails failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	response.OK(w, map[string]any{"emails": resp.GetEmails()})
}

// DeleteTreeAccessEmail revokes access for email from a tree.
// @Summary Delete tree access email
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param request body treeAccessEmailRequest true "Request body"
// @Success 200 {object} familyStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id}/access-emails [delete]
func (h *Handler) DeleteTreeAccessEmail(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	if strings.TrimSpace(treeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	var req treeAccessEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	err := h.client.DeleteTreeAccessEmail(r.Context(), treeID, req.Email)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("delete tree access email failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}

// ListPersons returns all persons in a tree.
// @Summary List persons
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Success 200 {object} personsSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id}/persons [get]
func (h *Handler) ListPersons(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	if strings.TrimSpace(treeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	resp, err := h.client.ListPersonsByTree(r.Context(), treeID)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("list persons failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	persons := make([]map[string]any, 0, len(resp.GetPersons()))
	for _, p := range resp.GetPersons() {
		persons = append(persons, toPersonJSON(p))
	}

	response.OK(w, map[string]any{"persons": persons})
}

// GetPerson returns person by ID in a tree.
// @Summary Get person
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param person_id path string true "Person ID"
// @Success 200 {object} personSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id}/persons/{person_id} [get]
func (h *Handler) GetPerson(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	if strings.TrimSpace(treeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	personID := chi.URLParam(r, "person_id")
	if strings.TrimSpace(personID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "person_id is required")
		return
	}

	resp, err := h.client.GetPerson(r.Context(), treeID, personID)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("get person failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	response.OK(w, map[string]any{"person": toPersonJSON(resp.GetPerson())})
}

// AddParent adds a parent for a child in the tree.
// @Summary Add parent
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param request body addParentRequest true "Request body"
// @Success 200 {object} addParentSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id}/parents [post]
func (h *Handler) AddParent(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	if strings.TrimSpace(treeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	var req addParentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	grpcReq := &familytreepb.AddParentRequest{
		TreeId:     treeID,
		ChildId:    req.ChildID,
		Role:       parseParentRole(req.Role),
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Patronymic: req.Patronymic,
	}

	resp, err := h.client.AddParent(r.Context(), grpcReq)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("add parent failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	data := map[string]any{"parent": toPersonJSON(resp.GetParent())}
	if resp.GetAutoCreatedSecondParent() != nil {
		data["auto_created_second_parent"] = toPersonJSON(resp.GetAutoCreatedSecondParent())
	}

	response.OK(w, data)
}

// AddChild adds a child for parents in the tree.
// @Summary Add child
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param request body addChildRequest true "Request body"
// @Success 200 {object} addChildSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id}/children [post]
func (h *Handler) AddChild(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	if strings.TrimSpace(treeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	var req addChildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	grpcReq := &familytreepb.AddChildRequest{
		TreeId:     treeID,
		Parent1Id:  req.Parent1ID,
		Parent2Id:  req.Parent2ID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Patronymic: req.Patronymic,
		Gender:     parseGender(req.Gender),
	}

	resp, err := h.client.AddChild(r.Context(), grpcReq)
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("add child failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	data := map[string]any{"child": toPersonJSON(resp.GetChild())}
	if resp.GetAutoCreatedParent() != nil {
		data["auto_created_parent"] = toPersonJSON(resp.GetAutoCreatedParent())
	}

	response.OK(w, data)
}

// AddPartner adds a partner for a person in the tree.
// @Summary Add partner
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param request body addPartnerRequest true "Request body"
// @Success 200 {object} addPartnerSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id}/partners [post]
func (h *Handler) AddPartner(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	if strings.TrimSpace(treeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	var req addPartnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	resp, err := h.client.AddPartner(r.Context(), &familytreepb.AddPartnerRequest{
		TreeId:     treeID,
		PersonId:   req.PersonID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Patronymic: req.Patronymic,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("add partner failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	response.OK(w, map[string]any{"partner": toPersonJSON(resp.GetPartner())})
}

// UpdatePersonName updates person name fields.
// @Summary Update person name
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param person_id path string true "Person ID"
// @Param request body updatePersonNameRequest true "Request body"
// @Success 200 {object} personSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id}/persons/{person_id} [patch]
func (h *Handler) UpdatePersonName(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	if strings.TrimSpace(treeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	personID := chi.URLParam(r, "person_id")
	if strings.TrimSpace(personID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "person_id is required")
		return
	}

	var req updatePersonNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	resp, err := h.client.UpdatePersonName(r.Context(), &familytreepb.UpdatePersonNameRequest{
		TreeId:     treeID,
		PersonId:   personID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Patronymic: req.Patronymic,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("update person name failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	response.OK(w, map[string]any{"person": toPersonJSON(resp.GetPerson())})
}

// DeletePerson deletes a person from tree.
// @Summary Delete person
// @Tags familytree
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tree_id path string true "Tree ID"
// @Param person_id path string true "Person ID"
// @Success 200 {object} familyStatusSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/familytree/{tree_id}/persons/{person_id} [delete]
func (h *Handler) DeletePerson(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "tree_id")
	if strings.TrimSpace(treeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	personID := chi.URLParam(r, "person_id")
	if strings.TrimSpace(personID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "person_id is required")
		return
	}

	err := h.client.DeletePersonInTree(r.Context(), &familytreepb.DeletePersonInTreeRequest{
		TreeId:   treeID,
		PersonId: personID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("delete person failed", slog.String("error", err.Error()))
		response.Error(w, status, "familytree_error", msg)
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}

func parseParentRole(v string) familytreepb.ParentRole {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "FATHER":
		return familytreepb.ParentRole_PARENT_ROLE_FATHER
	case "MOTHER":
		return familytreepb.ParentRole_PARENT_ROLE_MOTHER
	default:
		return familytreepb.ParentRole_PARENT_ROLE_UNSPECIFIED
	}
}

func parseGender(v string) familytreepb.Gender {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "MALE":
		return familytreepb.Gender_GENDER_MALE
	case "FEMALE":
		return familytreepb.Gender_GENDER_FEMALE
	default:
		return familytreepb.Gender_GENDER_UNSPECIFIED
	}
}

func toPersonJSON(p *familytreepb.Person) map[string]any {
	if p == nil {
		return map[string]any{}
	}

	return map[string]any{
		"id":              p.GetId(),
		"tree_id":         p.GetTreeId(),
		"first_name":      p.GetFirstName(),
		"last_name":       p.GetLastName(),
		"patronymic":      p.GetPatronymic(),
		"gender":          p.GetGender().String(),
		"avatar_photo_id": p.GetAvatarPhotoId(),
	}
}

func toTreeJSON(t *familytreepb.Tree) map[string]any {
	if t == nil {
		return map[string]any{}
	}

	return map[string]any{
		"id":                     t.GetId(),
		"creator_id":             t.GetCreatorId(),
		"created_at_unix":        t.GetCreatedAtUnix(),
		"is_view_restricted":     t.GetIsViewRestricted(),
		"is_public_on_main_page": t.GetIsPublicOnMainPage(),
	}
}
