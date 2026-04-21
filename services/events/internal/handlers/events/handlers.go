package events

import (
	"context"

	eventspb "github.com/Bad-Utya/myforebears-backend/gen/go/events"
	"github.com/Bad-Utya/myforebears-backend/services/events/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/events/internal/lib/grpcerr"
	"google.golang.org/grpc"
)

type EventsService interface {
	CreateEventType(ctx context.Context, requestUserID int, name string, mode models.PrimaryPersonsMode, count int) (models.EventType, error)
	GetEventType(ctx context.Context, requestUserID int, eventTypeID string) (models.EventType, error)
	ListEventTypes(ctx context.Context, requestUserID int) ([]models.EventType, error)
	DeleteEventType(ctx context.Context, requestUserID int, eventTypeID string) error
	CreateEvent(ctx context.Context, requestUserID int, treeID string, eventTypeID string, primaryPersonIDs []string, additionalPersonIDs []string, dateISO string, datePrecision models.EventDatePrecision, dateBound models.EventDateBound) (models.Event, error)
	GetEvent(ctx context.Context, requestUserID int, eventID string) (models.Event, error)
	ListEventsByTree(ctx context.Context, requestUserID int, treeID string) ([]models.Event, error)
	UpdateEvent(ctx context.Context, requestUserID int, eventID string, eventTypeID string, primaryPersonIDs []string, additionalPersonIDs []string, dateISO string, datePrecision models.EventDatePrecision, dateBound models.EventDateBound) (models.Event, error)
	DeleteEvent(ctx context.Context, requestUserID int, eventID string) error
}

type Handler struct {
	eventspb.UnimplementedEventsServiceServer
	service EventsService
}

func New(service EventsService) *Handler {
	return &Handler{service: service}
}

func Register(gRPC *grpc.Server, service EventsService) {
	eventspb.RegisterEventsServiceServer(gRPC, New(service))
}

func (h *Handler) CreateEventType(ctx context.Context, req *eventspb.CreateEventTypeRequest) (*eventspb.CreateEventTypeResponse, error) {
	eventType, err := h.service.CreateEventType(
		ctx,
		int(req.GetRequestUserId()),
		req.GetName(),
		toModelPrimaryPersonsMode(req.GetPrimaryPersonsMode()),
		int(req.GetPrimaryPersonsCount()),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &eventspb.CreateEventTypeResponse{EventType: toProtoEventType(eventType)}, nil
}

func (h *Handler) GetEventType(ctx context.Context, req *eventspb.GetEventTypeRequest) (*eventspb.GetEventTypeResponse, error) {
	eventType, err := h.service.GetEventType(ctx, int(req.GetRequestUserId()), req.GetEventTypeId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &eventspb.GetEventTypeResponse{EventType: toProtoEventType(eventType)}, nil
}

func (h *Handler) ListEventTypes(ctx context.Context, req *eventspb.ListEventTypesRequest) (*eventspb.ListEventTypesResponse, error) {
	eventTypes, err := h.service.ListEventTypes(ctx, int(req.GetRequestUserId()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	out := make([]*eventspb.EventType, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		out = append(out, toProtoEventType(eventType))
	}

	return &eventspb.ListEventTypesResponse{EventTypes: out}, nil
}

func (h *Handler) DeleteEventType(ctx context.Context, req *eventspb.DeleteEventTypeRequest) (*eventspb.DeleteEventTypeResponse, error) {
	err := h.service.DeleteEventType(ctx, int(req.GetRequestUserId()), req.GetEventTypeId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &eventspb.DeleteEventTypeResponse{}, nil
}

func (h *Handler) CreateEvent(ctx context.Context, req *eventspb.CreateEventRequest) (*eventspb.CreateEventResponse, error) {
	event, err := h.service.CreateEvent(
		ctx,
		int(req.GetRequestUserId()),
		req.GetTreeId(),
		req.GetEventTypeId(),
		req.GetPrimaryPersonIds(),
		req.GetAdditionalPersonIds(),
		req.GetDateIso(),
		toModelDatePrecision(req.GetDatePrecision()),
		toModelDateBound(req.GetDateBound()),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &eventspb.CreateEventResponse{Event: toProtoEvent(event)}, nil
}

func (h *Handler) GetEvent(ctx context.Context, req *eventspb.GetEventRequest) (*eventspb.GetEventResponse, error) {
	event, err := h.service.GetEvent(ctx, int(req.GetRequestUserId()), req.GetEventId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &eventspb.GetEventResponse{Event: toProtoEvent(event)}, nil
}

func (h *Handler) ListEventsByTree(ctx context.Context, req *eventspb.ListEventsByTreeRequest) (*eventspb.ListEventsByTreeResponse, error) {
	events, err := h.service.ListEventsByTree(ctx, int(req.GetRequestUserId()), req.GetTreeId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	out := make([]*eventspb.Event, 0, len(events))
	for _, event := range events {
		out = append(out, toProtoEvent(event))
	}

	return &eventspb.ListEventsByTreeResponse{Events: out}, nil
}

func (h *Handler) UpdateEvent(ctx context.Context, req *eventspb.UpdateEventRequest) (*eventspb.UpdateEventResponse, error) {
	event, err := h.service.UpdateEvent(
		ctx,
		int(req.GetRequestUserId()),
		req.GetEventId(),
		req.GetEventTypeId(),
		req.GetPrimaryPersonIds(),
		req.GetAdditionalPersonIds(),
		req.GetDateIso(),
		toModelDatePrecision(req.GetDatePrecision()),
		toModelDateBound(req.GetDateBound()),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &eventspb.UpdateEventResponse{Event: toProtoEvent(event)}, nil
}

func (h *Handler) DeleteEvent(ctx context.Context, req *eventspb.DeleteEventRequest) (*eventspb.DeleteEventResponse, error) {
	err := h.service.DeleteEvent(ctx, int(req.GetRequestUserId()), req.GetEventId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &eventspb.DeleteEventResponse{}, nil
}

func toModelPrimaryPersonsMode(mode eventspb.PrimaryPersonsMode) models.PrimaryPersonsMode {
	switch mode {
	case eventspb.PrimaryPersonsMode_PRIMARY_PERSONS_MODE_FIXED:
		return models.PrimaryPersonsModeFixed
	case eventspb.PrimaryPersonsMode_PRIMARY_PERSONS_MODE_UNLIMITED:
		return models.PrimaryPersonsModeUnlimited
	default:
		return ""
	}
}

func toModelDatePrecision(precision eventspb.EventDatePrecision) models.EventDatePrecision {
	switch precision {
	case eventspb.EventDatePrecision_EVENT_DATE_PRECISION_DAY:
		return models.EventDatePrecisionDay
	case eventspb.EventDatePrecision_EVENT_DATE_PRECISION_MONTH:
		return models.EventDatePrecisionMonth
	case eventspb.EventDatePrecision_EVENT_DATE_PRECISION_YEAR:
		return models.EventDatePrecisionYear
	default:
		return ""
	}
}

func toModelDateBound(bound eventspb.EventDateBound) models.EventDateBound {
	switch bound {
	case eventspb.EventDateBound_EVENT_DATE_BOUND_EXACT:
		return models.EventDateBoundExact
	case eventspb.EventDateBound_EVENT_DATE_BOUND_NOT_BEFORE:
		return models.EventDateBoundNotBefore
	case eventspb.EventDateBound_EVENT_DATE_BOUND_NOT_AFTER:
		return models.EventDateBoundNotAfter
	default:
		return ""
	}
}

func toProtoPrimaryPersonsMode(mode models.PrimaryPersonsMode) eventspb.PrimaryPersonsMode {
	switch mode {
	case models.PrimaryPersonsModeFixed:
		return eventspb.PrimaryPersonsMode_PRIMARY_PERSONS_MODE_FIXED
	case models.PrimaryPersonsModeUnlimited:
		return eventspb.PrimaryPersonsMode_PRIMARY_PERSONS_MODE_UNLIMITED
	default:
		return eventspb.PrimaryPersonsMode_PRIMARY_PERSONS_MODE_UNSPECIFIED
	}
}

func toProtoDatePrecision(precision models.EventDatePrecision) eventspb.EventDatePrecision {
	switch precision {
	case models.EventDatePrecisionDay:
		return eventspb.EventDatePrecision_EVENT_DATE_PRECISION_DAY
	case models.EventDatePrecisionMonth:
		return eventspb.EventDatePrecision_EVENT_DATE_PRECISION_MONTH
	case models.EventDatePrecisionYear:
		return eventspb.EventDatePrecision_EVENT_DATE_PRECISION_YEAR
	default:
		return eventspb.EventDatePrecision_EVENT_DATE_PRECISION_UNSPECIFIED
	}
}

func toProtoDateBound(bound models.EventDateBound) eventspb.EventDateBound {
	switch bound {
	case models.EventDateBoundExact:
		return eventspb.EventDateBound_EVENT_DATE_BOUND_EXACT
	case models.EventDateBoundNotBefore:
		return eventspb.EventDateBound_EVENT_DATE_BOUND_NOT_BEFORE
	case models.EventDateBoundNotAfter:
		return eventspb.EventDateBound_EVENT_DATE_BOUND_NOT_AFTER
	default:
		return eventspb.EventDateBound_EVENT_DATE_BOUND_UNSPECIFIED
	}
}

func toProtoEventType(eventType models.EventType) *eventspb.EventType {
	return &eventspb.EventType{
		Id:                  eventType.ID.String(),
		OwnerUserId:         int32(eventType.OwnerUserID),
		IsSystem:            eventType.IsSystem,
		Name:                eventType.Name,
		PrimaryPersonsMode:  toProtoPrimaryPersonsMode(eventType.PrimaryPersonsMode),
		PrimaryPersonsCount: int32(eventType.PrimaryPersonsCount),
	}
}

func toProtoEvent(event models.Event) *eventspb.Event {
	primaryIDs := make([]string, 0, len(event.PrimaryPersonIDs))
	for _, id := range event.PrimaryPersonIDs {
		primaryIDs = append(primaryIDs, id.String())
	}

	additionalIDs := make([]string, 0, len(event.AdditionalPersonIDs))
	for _, id := range event.AdditionalPersonIDs {
		additionalIDs = append(additionalIDs, id.String())
	}

	return &eventspb.Event{
		Id:                  event.ID.String(),
		TreeId:              event.TreeID.String(),
		EventTypeId:         event.EventTypeID.String(),
		PrimaryPersonIds:    primaryIDs,
		AdditionalPersonIds: additionalIDs,
		DateIso:             event.DateValue.Format("2006-01-02"),
		DatePrecision:       toProtoDatePrecision(event.DatePrecision),
		DateBound:           toProtoDateBound(event.DateBound),
		CreatedAtUnix:       event.CreatedAt.Unix(),
		UpdatedAtUnix:       event.UpdatedAt.Unix(),
	}
}
