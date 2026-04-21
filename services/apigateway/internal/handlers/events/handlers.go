package events

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	eventspb "github.com/Bad-Utya/myforebears-backend/gen/go/events"
	eventsclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/events"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/grpcerr"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/response"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	log    *slog.Logger
	client *eventsclient.Client
}

func New(log *slog.Logger, client *eventsclient.Client) *Handler {
	return &Handler{log: log, client: client}
}

type createEventTypeRequest struct {
	Name                string `json:"name"`
	PrimaryPersonsMode  string `json:"primary_persons_mode"`
	PrimaryPersonsCount int    `json:"primary_persons_count"`
}

type upsertEventRequest struct {
	TreeID              string   `json:"tree_id"`
	EventTypeID         string   `json:"event_type_id"`
	PrimaryPersonIDs    []string `json:"primary_person_ids"`
	AdditionalPersonIDs []string `json:"additional_person_ids"`
	DateISO             string   `json:"date_iso"`
	DatePrecision       string   `json:"date_precision"`
	DateBound           string   `json:"date_bound"`
}

func (h *Handler) CreateEventType(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	var req createEventTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	resp, err := h.client.CreateEventType(r.Context(), &eventspb.CreateEventTypeRequest{
		RequestUserId:       int32(userID),
		Name:                req.Name,
		PrimaryPersonsMode:  parsePrimaryPersonsMode(req.PrimaryPersonsMode),
		PrimaryPersonsCount: int32(req.PrimaryPersonsCount),
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("create event type failed", slog.String("error", err.Error()))
		response.Error(w, status, "events_error", msg)
		return
	}

	response.OK(w, map[string]any{"event_type": toEventTypeJSON(resp.GetEventType())})
}

func (h *Handler) GetEventType(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	eventTypeID := chi.URLParam(r, "event_type_id")
	if strings.TrimSpace(eventTypeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "event_type_id is required")
		return
	}

	resp, err := h.client.GetEventType(r.Context(), &eventspb.GetEventTypeRequest{
		RequestUserId: int32(userID),
		EventTypeId:   eventTypeID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("get event type failed", slog.String("error", err.Error()))
		response.Error(w, status, "events_error", msg)
		return
	}

	response.OK(w, map[string]any{"event_type": toEventTypeJSON(resp.GetEventType())})
}

func (h *Handler) ListEventTypes(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	resp, err := h.client.ListEventTypes(r.Context(), &eventspb.ListEventTypesRequest{RequestUserId: int32(userID)})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("list event types failed", slog.String("error", err.Error()))
		response.Error(w, status, "events_error", msg)
		return
	}

	eventTypes := make([]map[string]any, 0, len(resp.GetEventTypes()))
	for _, eventType := range resp.GetEventTypes() {
		eventTypes = append(eventTypes, toEventTypeJSON(eventType))
	}

	response.OK(w, map[string]any{"event_types": eventTypes})
}

func (h *Handler) DeleteEventType(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	eventTypeID := chi.URLParam(r, "event_type_id")
	if strings.TrimSpace(eventTypeID) == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "event_type_id is required")
		return
	}

	err = h.client.DeleteEventType(r.Context(), &eventspb.DeleteEventTypeRequest{
		RequestUserId: int32(userID),
		EventTypeId:   eventTypeID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("delete event type failed", slog.String("error", err.Error()))
		response.Error(w, status, "events_error", msg)
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	var req upsertEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	resp, err := h.client.CreateEvent(r.Context(), &eventspb.CreateEventRequest{
		RequestUserId:       int32(userID),
		TreeId:              req.TreeID,
		EventTypeId:         req.EventTypeID,
		PrimaryPersonIds:    req.PrimaryPersonIDs,
		AdditionalPersonIds: req.AdditionalPersonIDs,
		DateIso:             req.DateISO,
		DatePrecision:       parseDatePrecision(req.DatePrecision),
		DateBound:           parseDateBound(req.DateBound),
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("create event failed", slog.String("error", err.Error()))
		response.Error(w, status, "events_error", msg)
		return
	}

	response.OK(w, map[string]any{"event": toEventJSON(resp.GetEvent())})
}

func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.GetEvent(r.Context(), &eventspb.GetEventRequest{
		RequestUserId: int32(userID),
		EventId:       eventID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("get event failed", slog.String("error", err.Error()))
		response.Error(w, status, "events_error", msg)
		return
	}

	response.OK(w, map[string]any{"event": toEventJSON(resp.GetEvent())})
}

func (h *Handler) ListEventsByTree(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
		return
	}

	treeID := strings.TrimSpace(r.URL.Query().Get("tree_id"))
	if treeID == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "tree_id is required")
		return
	}

	resp, err := h.client.ListEventsByTree(r.Context(), &eventspb.ListEventsByTreeRequest{
		RequestUserId: int32(userID),
		TreeId:        treeID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("list events by tree failed", slog.String("error", err.Error()))
		response.Error(w, status, "events_error", msg)
		return
	}

	events := make([]map[string]any, 0, len(resp.GetEvents()))
	for _, event := range resp.GetEvents() {
		events = append(events, toEventJSON(event))
	}

	response.OK(w, map[string]any{"events": events})
}

func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
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

	var req upsertEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	resp, err := h.client.UpdateEvent(r.Context(), &eventspb.UpdateEventRequest{
		RequestUserId:       int32(userID),
		EventId:             eventID,
		EventTypeId:         req.EventTypeID,
		PrimaryPersonIds:    req.PrimaryPersonIDs,
		AdditionalPersonIds: req.AdditionalPersonIDs,
		DateIso:             req.DateISO,
		DatePrecision:       parseDatePrecision(req.DatePrecision),
		DateBound:           parseDateBound(req.DateBound),
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("update event failed", slog.String("error", err.Error()))
		response.Error(w, status, "events_error", msg)
		return
	}

	response.OK(w, map[string]any{"event": toEventJSON(resp.GetEvent())})
}

func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
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

	err = h.client.DeleteEvent(r.Context(), &eventspb.DeleteEventRequest{
		RequestUserId: int32(userID),
		EventId:       eventID,
	})
	if err != nil {
		status, msg := grpcerr.HTTPStatus(err)
		h.log.Error("delete event failed", slog.String("error", err.Error()))
		response.Error(w, status, "events_error", msg)
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}

func parsePrimaryPersonsMode(v string) eventspb.PrimaryPersonsMode {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "FIXED":
		return eventspb.PrimaryPersonsMode_PRIMARY_PERSONS_MODE_FIXED
	case "UNLIMITED":
		return eventspb.PrimaryPersonsMode_PRIMARY_PERSONS_MODE_UNLIMITED
	default:
		return eventspb.PrimaryPersonsMode_PRIMARY_PERSONS_MODE_UNSPECIFIED
	}
}

func parseDatePrecision(v string) eventspb.EventDatePrecision {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "DAY":
		return eventspb.EventDatePrecision_EVENT_DATE_PRECISION_DAY
	case "MONTH":
		return eventspb.EventDatePrecision_EVENT_DATE_PRECISION_MONTH
	case "YEAR":
		return eventspb.EventDatePrecision_EVENT_DATE_PRECISION_YEAR
	default:
		return eventspb.EventDatePrecision_EVENT_DATE_PRECISION_UNSPECIFIED
	}
}

func parseDateBound(v string) eventspb.EventDateBound {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "EXACT":
		return eventspb.EventDateBound_EVENT_DATE_BOUND_EXACT
	case "NOT_BEFORE":
		return eventspb.EventDateBound_EVENT_DATE_BOUND_NOT_BEFORE
	case "NOT_AFTER":
		return eventspb.EventDateBound_EVENT_DATE_BOUND_NOT_AFTER
	default:
		return eventspb.EventDateBound_EVENT_DATE_BOUND_UNSPECIFIED
	}
}

func toEventTypeJSON(eventType *eventspb.EventType) map[string]any {
	if eventType == nil {
		return map[string]any{}
	}

	return map[string]any{
		"id":                    eventType.GetId(),
		"owner_user_id":         eventType.GetOwnerUserId(),
		"is_system":             eventType.GetIsSystem(),
		"name":                  eventType.GetName(),
		"primary_persons_mode":  eventType.GetPrimaryPersonsMode().String(),
		"primary_persons_count": eventType.GetPrimaryPersonsCount(),
	}
}

func toEventJSON(event *eventspb.Event) map[string]any {
	if event == nil {
		return map[string]any{}
	}

	return map[string]any{
		"id":                    event.GetId(),
		"tree_id":               event.GetTreeId(),
		"event_type_id":         event.GetEventTypeId(),
		"primary_person_ids":    event.GetPrimaryPersonIds(),
		"additional_person_ids": event.GetAdditionalPersonIds(),
		"date_iso":              event.GetDateIso(),
		"date_precision":        event.GetDatePrecision().String(),
		"date_bound":            event.GetDateBound().String(),
		"created_at_unix":       event.GetCreatedAtUnix(),
		"updated_at_unix":       event.GetUpdatedAtUnix(),
	}
}
