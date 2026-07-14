package familytree

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	eventspb "github.com/Bad-Utya/myforebears-backend/gen/go/events"
	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	photospb "github.com/Bad-Utya/myforebears-backend/gen/go/photos"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/grpcerr"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/response"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type exportPublicPersonRequest struct {
	TreeID   string `json:"tree_id"`
	PersonID string `json:"person_id"`
}
type importPublicPersonRequest struct {
	TreeID           string `json:"tree_id"`
	AttachToPersonID string `json:"attach_to_person_id"`
	Attachment       string `json:"attachment"`
}
type createTreeFromPublicRequest struct {
	TreeName string `json:"tree_name"`
}
type publicEventInput struct {
	ID            string `json:"id"`
	EventTypeID   string `json:"event_type_id"`
	EventTypeName string `json:"event_type_name"`
	DateISO       string `json:"date_iso"`
	DatePrecision string `json:"date_precision"`
	DateBound     string `json:"date_bound"`
	DateUnknown   bool   `json:"date_unknown"`
}
type updatePublicPersonRequest struct {
	FirstName  string             `json:"first_name"`
	LastName   string             `json:"last_name"`
	Patronymic string             `json:"patronymic"`
	Gender     string             `json:"gender"`
	Biography  string             `json:"biography"`
	Events     []publicEventInput `json:"events"`
}

func (h *Handler) CreatePublicPerson(w http.ResponseWriter, r *http.Request) {
	userID, ok := publicUserID(w, r)
	if !ok {
		return
	}
	resp, err := h.client.CreatePublicPerson(r.Context(), userID)
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"person": publicPersonJSON(resp.GetPerson())})
}

func (h *Handler) ExportPersonToPublic(w http.ResponseWriter, r *http.Request) {
	userID, ok := publicUserID(w, r)
	if !ok {
		return
	}
	var req exportPublicPersonRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		response.Error(w, 400, "bad_request", "invalid request body")
		return
	}
	treeAccess, err := h.client.GetTreeAccessInfo(r.Context(), req.TreeID)
	if err != nil {
		publicError(w, err)
		return
	}
	if int(treeAccess.GetTree().GetCreatorId()) != userID {
		response.Error(w, http.StatusForbidden, "forbidden", "only the tree owner can publish a person")
		return
	}
	personResp, err := h.client.GetPerson(r.Context(), req.TreeID, req.PersonID)
	if err != nil {
		publicError(w, err)
		return
	}
	person := personResp.GetPerson()
	eventsResp, err := h.eventsClient.ListEventsByTree(r.Context(), &eventspb.ListEventsByTreeRequest{TreeId: req.TreeID})
	if err != nil {
		publicError(w, err)
		return
	}
	inputs := make([]*familytreepb.PublicPersonEventInput, 0)
	for _, event := range eventsResp.GetEvents() {
		if !containsString(event.GetPrimaryPersonIds(), req.PersonID) && !containsString(event.GetAdditionalPersonIds(), req.PersonID) {
			continue
		}
		et, err := h.eventsClient.GetEventType(r.Context(), &eventspb.GetEventTypeRequest{RequestUserId: int32(userID), EventTypeId: event.GetEventTypeId()})
		if err != nil {
			publicError(w, err)
			return
		}
		inputs = append(inputs, &familytreepb.PublicPersonEventInput{SourceEventId: event.GetId(), EventTypeId: event.GetEventTypeId(), EventTypeName: et.GetEventType().GetName(), DateIso: event.GetDateIso(), DatePrecision: event.GetDatePrecision().String(), DateBound: event.GetDateBound().String(), DateUnknown: event.GetDateUnknown()})
	}
	created, err := h.client.CreatePublicPersonSnapshot(r.Context(), &familytreepb.CreatePublicPersonSnapshotRequest{RequestUserId: int32(userID), FirstName: person.GetFirstName(), LastName: person.GetLastName(), Patronymic: person.GetPatronymic(), Gender: person.GetGender(), Biography: person.GetBiography(), Events: inputs})
	if err != nil {
		publicError(w, err)
		return
	}
	mappings := make([]*photospb.EventPhotoMapping, 0, len(inputs))
	for i, event := range created.GetPerson().GetEvents() {
		if i < len(inputs) {
			mappings = append(mappings, &photospb.EventPhotoMapping{SourceEventId: inputs[i].GetSourceEventId(), TargetEventId: event.GetId()})
		}
	}
	_, err = h.photosClient.CopyPersonMediaToPublic(r.Context(), &photospb.CopyPersonMediaToPublicRequest{RequestUserId: int32(userID), TreeId: req.TreeID, PersonId: req.PersonID, PublicPersonId: created.GetPerson().GetId(), EventMappings: mappings})
	if err != nil {
		_ = h.client.DeletePublicPerson(r.Context(), userID, created.GetPerson().GetId())
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"person": publicPersonJSON(created.GetPerson())})
}

func (h *Handler) GetPublicPerson(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.GetPublicPerson(r.Context(), chi.URLParam(r, "public_person_id"))
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"person": publicPersonJSON(resp.GetPerson())})
}
func (h *Handler) ListRandomPublicPersons(w http.ResponseWriter, r *http.Request) {
	limit := queryLimit(r, 20)
	resp, err := h.client.ListRandomPublicPersons(r.Context(), limit)
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"persons": publicPersonsJSON(resp.GetPersons())})
}
func (h *Handler) SearchPublicPersons(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.SearchPublicPersons(r.Context(), r.URL.Query().Get("q"), selectedTagCodes(r), queryLimit(r, 20))
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"persons": publicPersonsJSON(resp.GetPersons())})
}
func (h *Handler) ListPublicPersonsByOwner(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "user_id"))
	if err != nil {
		response.Error(w, 400, "bad_request", "invalid user_id")
		return
	}
	resp, err := h.client.ListPublicPersonsByOwner(r.Context(), id, queryLimit(r, 20))
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"persons": publicPersonsJSON(resp.GetPersons())})
}

func (h *Handler) UpdatePublicPerson(w http.ResponseWriter, r *http.Request) {
	userID, ok := publicUserID(w, r)
	if !ok {
		return
	}
	var body updatePublicPersonRequest
	if json.NewDecoder(r.Body).Decode(&body) != nil {
		response.Error(w, 400, "bad_request", "invalid request body")
		return
	}
	events := make([]*familytreepb.PublicPersonEventInput, 0, len(body.Events))
	for _, e := range body.Events {
		events = append(events, &familytreepb.PublicPersonEventInput{Id: e.ID, EventTypeId: e.EventTypeID, EventTypeName: e.EventTypeName, DateIso: e.DateISO, DatePrecision: e.DatePrecision, DateBound: e.DateBound, DateUnknown: e.DateUnknown})
	}
	resp, err := h.client.UpdatePublicPerson(r.Context(), &familytreepb.UpdatePublicPersonRequest{RequestUserId: int32(userID), PublicPersonId: chi.URLParam(r, "public_person_id"), FirstName: body.FirstName, LastName: body.LastName, Patronymic: body.Patronymic, Gender: parseGender(body.Gender), Biography: body.Biography, Events: events})
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"person": publicPersonJSON(resp.GetPerson())})
}

func (h *Handler) DeletePublicPerson(w http.ResponseWriter, r *http.Request) {
	userID, ok := publicUserID(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, "public_person_id")
	if err := h.photosClient.DeletePublicPersonMedia(r.Context(), userID, id); err != nil {
		publicError(w, err)
		return
	}
	if err := h.client.DeletePublicPerson(r.Context(), userID, id); err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}

func (h *Handler) ImportPublicPerson(w http.ResponseWriter, r *http.Request) {
	userID, ok := publicUserID(w, r)
	if !ok {
		return
	}
	publicID := chi.URLParam(r, "public_person_id")
	var body importPublicPersonRequest
	if json.NewDecoder(r.Body).Decode(&body) != nil {
		response.Error(w, 400, "bad_request", "invalid request body")
		return
	}
	pub, err := h.client.GetPublicPerson(r.Context(), publicID)
	if err != nil {
		publicError(w, err)
		return
	}
	core, err := h.client.ImportPublicPersonIntoTree(r.Context(), &familytreepb.ImportPublicPersonIntoTreeRequest{RequestUserId: int32(userID), PublicPersonId: publicID, TreeId: body.TreeID, AttachToPersonId: body.AttachToPersonID, Attachment: parseAttachment(body.Attachment)})
	if err != nil {
		publicError(w, err)
		return
	}
	mappings, err := h.importPublicEvents(r, userID, body.TreeID, core.GetPerson().GetId(), pub.GetPerson().GetEvents())
	if err != nil {
		publicError(w, err)
		return
	}
	_, err = h.photosClient.CopyPublicPersonMediaToTree(r.Context(), &photospb.CopyPublicPersonMediaToTreeRequest{RequestUserId: int32(userID), PublicPersonId: publicID, TreeId: body.TreeID, PersonId: core.GetPerson().GetId(), EventMappings: mappings})
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"person": toPersonJSON(core.GetPerson())})
}

func (h *Handler) CreateTreeFromPublicPerson(w http.ResponseWriter, r *http.Request) {
	userID, ok := publicUserID(w, r)
	if !ok {
		return
	}
	publicID := chi.URLParam(r, "public_person_id")
	var body createTreeFromPublicRequest
	_ = json.NewDecoder(r.Body).Decode(&body)
	pub, err := h.client.GetPublicPerson(r.Context(), publicID)
	if err != nil {
		publicError(w, err)
		return
	}
	core, err := h.client.CreateTreeFromPublicPerson(r.Context(), &familytreepb.CreateTreeFromPublicPersonRequest{RequestUserId: int32(userID), PublicPersonId: publicID, TreeName: body.TreeName})
	if err != nil {
		publicError(w, err)
		return
	}
	mappings, err := h.importPublicEvents(r, userID, core.GetTree().GetId(), core.GetPerson().GetId(), pub.GetPerson().GetEvents())
	if err != nil {
		publicError(w, err)
		return
	}
	_, err = h.photosClient.CopyPublicPersonMediaToTree(r.Context(), &photospb.CopyPublicPersonMediaToTreeRequest{RequestUserId: int32(userID), PublicPersonId: publicID, TreeId: core.GetTree().GetId(), PersonId: core.GetPerson().GetId(), EventMappings: mappings})
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"tree": toTreeJSON(core.GetTree()), "person": toPersonJSON(core.GetPerson())})
}

func (h *Handler) UploadPublicPersonPhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := publicUserID(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(16 << 20); err != nil {
		response.Error(w, 400, "bad_request", "invalid multipart body")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, 400, "bad_request", "file is required")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		response.Error(w, 400, "bad_request", "failed to read file")
		return
	}
	mime := header.Header.Get("Content-Type")
	resp, err := h.photosClient.UploadPublicPersonPhoto(r.Context(), &photospb.UploadPublicPersonPhotoRequest{RequestUserId: int32(userID), PublicPersonId: chi.URLParam(r, "public_person_id"), FileName: header.Filename, MimeType: mime, Content: data, IsAvatar: r.FormValue("is_avatar") == "true"})
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"photo": resp.GetPhoto()})
}
func (h *Handler) ListPublicPersonPhotos(w http.ResponseWriter, r *http.Request) {
	resp, err := h.photosClient.ListPublicPersonPhotos(r.Context(), chi.URLParam(r, "public_person_id"))
	if err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]any{"photos": resp.GetPhotos()})
}
func (h *Handler) GetPublicPersonPhoto(w http.ResponseWriter, r *http.Request) {
	resp, err := h.photosClient.GetPublicPersonPhoto(r.Context(), chi.URLParam(r, "public_person_id"), chi.URLParam(r, "photo_id"))
	if err != nil {
		publicError(w, err)
		return
	}
	w.Header().Set("Content-Type", resp.GetPhoto().GetMimeType())
	w.WriteHeader(200)
	_, _ = w.Write(resp.GetContent())
}
func (h *Handler) DeletePublicPersonPhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := publicUserID(w, r)
	if !ok {
		return
	}
	if err := h.photosClient.DeletePublicPersonPhoto(r.Context(), userID, chi.URLParam(r, "public_person_id"), chi.URLParam(r, "photo_id")); err != nil {
		publicError(w, err)
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}

func (h *Handler) importPublicEvents(r *http.Request, userID int, treeID, personID string, items []*familytreepb.PublicPersonEvent) ([]*photospb.EventPhotoMapping, error) {
	types, err := h.eventsClient.ListEventTypes(r.Context(), &eventspb.ListEventTypesRequest{RequestUserId: int32(userID)})
	if err != nil {
		return nil, err
	}
	byID := map[string]*eventspb.EventType{}
	byName := map[string]*eventspb.EventType{}
	for _, t := range types.GetEventTypes() {
		byID[t.GetId()] = t
		byName[strings.ToLower(t.GetName())] = t
	}
	out := make([]*photospb.EventPhotoMapping, 0, len(items))
	for _, item := range items {
		typ := byID[item.GetEventTypeId()]
		if typ == nil {
			typ = byName[strings.ToLower(item.GetEventTypeName())]
		}
		if typ == nil {
			created, err := h.eventsClient.CreateEventType(r.Context(), &eventspb.CreateEventTypeRequest{RequestUserId: int32(userID), Name: item.GetEventTypeName(), PrimaryPersonsMode: eventspb.PrimaryPersonsMode_PRIMARY_PERSONS_MODE_FIXED, PrimaryPersonsCount: 1})
			if err != nil {
				return nil, err
			}
			typ = created.GetEventType()
			byName[strings.ToLower(typ.GetName())] = typ
		}
		created, err := h.eventsClient.CreateEvent(r.Context(), &eventspb.CreateEventRequest{TreeId: treeID, EventTypeId: typ.GetId(), PrimaryPersonIds: []string{personID}, DateIso: item.GetDateIso(), DatePrecision: parseDatePrecision(item.GetDatePrecision()), DateBound: parseDateBound(item.GetDateBound()), DateUnknown: item.GetDateUnknown()})
		if err != nil {
			return nil, err
		}
		out = append(out, &photospb.EventPhotoMapping{SourceEventId: item.GetId(), TargetEventId: created.GetEvent().GetId()})
	}
	return out, nil
}

func publicUserID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, 401, "unauthorized", "invalid token claims")
		return 0, false
	}
	return id, true
}
func publicError(w http.ResponseWriter, err error) {
	status, msg := grpcerr.HTTPStatus(err)
	response.Error(w, status, "public_person_error", msg)
}
func queryLimit(r *http.Request, fallback int) int {
	v, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if v <= 0 {
		return fallback
	}
	return v
}
func containsString(items []string, want string) bool {
	for _, v := range items {
		if v == want {
			return true
		}
	}
	return false
}
func parseAttachment(v string) familytreepb.PublicPersonAttachment {
	switch strings.ToUpper(v) {
	case "PARENT":
		return familytreepb.PublicPersonAttachment_PUBLIC_PERSON_ATTACHMENT_AS_PARENT
	case "CHILD":
		return familytreepb.PublicPersonAttachment_PUBLIC_PERSON_ATTACHMENT_AS_CHILD
	case "PARTNER":
		return familytreepb.PublicPersonAttachment_PUBLIC_PERSON_ATTACHMENT_AS_PARTNER
	}
	return familytreepb.PublicPersonAttachment_PUBLIC_PERSON_ATTACHMENT_UNSPECIFIED
}
func parseDatePrecision(v string) eventspb.EventDatePrecision {
	if strings.Contains(v, "MONTH") {
		return eventspb.EventDatePrecision_EVENT_DATE_PRECISION_MONTH
	}
	if strings.Contains(v, "YEAR") {
		return eventspb.EventDatePrecision_EVENT_DATE_PRECISION_YEAR
	}
	return eventspb.EventDatePrecision_EVENT_DATE_PRECISION_DAY
}
func parseDateBound(v string) eventspb.EventDateBound {
	if strings.Contains(v, "NOT_BEFORE") {
		return eventspb.EventDateBound_EVENT_DATE_BOUND_NOT_BEFORE
	}
	if strings.Contains(v, "NOT_AFTER") {
		return eventspb.EventDateBound_EVENT_DATE_BOUND_NOT_AFTER
	}
	return eventspb.EventDateBound_EVENT_DATE_BOUND_EXACT
}
func publicPersonJSON(p *familytreepb.PublicPerson) any        { return p }
func publicPersonsJSON(items []*familytreepb.PublicPerson) any { return items }
