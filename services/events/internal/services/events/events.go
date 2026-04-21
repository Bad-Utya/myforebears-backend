package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Bad-Utya/myforebears-backend/services/events/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/events/internal/storage"
	"github.com/google/uuid"
)

var (
	ErrInvalidEventTypeID            = errors.New("invalid event type id")
	ErrInvalidEventID                = errors.New("invalid event id")
	ErrInvalidTreeID                 = errors.New("invalid tree id")
	ErrInvalidUserID                 = errors.New("invalid user id")
	ErrInvalidEventTypeName          = errors.New("event type name is required")
	ErrInvalidPrimaryPersonsMode     = errors.New("invalid primary persons mode")
	ErrInvalidPrimaryPersonsCount    = errors.New("invalid primary persons count")
	ErrInvalidEventDate              = errors.New("invalid event date")
	ErrInvalidEventDatePrecision     = errors.New("invalid event date precision")
	ErrInvalidEventDateBound         = errors.New("invalid event date bound")
	ErrInvalidPrimaryPersons         = errors.New("invalid primary persons for event type")
	ErrDuplicatePersonInParticipants = errors.New("duplicate person in participants")
	ErrParticipantListsOverlap       = errors.New("person cannot be both primary and additional participant")
	ErrEventTypeNotFound             = errors.New("event type not found")
	ErrEventNotFound                 = errors.New("event not found")
	ErrForbidden                     = errors.New("forbidden")
	ErrCannotDeleteSystemEventType   = errors.New("system event type cannot be deleted")
	ErrEventTypeInUse                = errors.New("event type is used by existing events")
	ErrEventTypeAlreadyExists        = errors.New("event type with this name already exists")
)

type FamilyTreeValidator interface {
	ValidatePersonsInTree(ctx context.Context, requestUserID int, treeID string, personIDs []string) error
}

type Service struct {
	log        *slog.Logger
	storage    storage.Storage
	familyTree FamilyTreeValidator
}

func New(log *slog.Logger, storage storage.Storage, familyTree FamilyTreeValidator) *Service {
	return &Service{log: log, storage: storage, familyTree: familyTree}
}

func (s *Service) CreateEventType(ctx context.Context, requestUserID int, name string, mode models.PrimaryPersonsMode, count int) (models.EventType, error) {
	const op = "service.events.CreateEventType"

	if requestUserID <= 0 {
		return models.EventType{}, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return models.EventType{}, fmt.Errorf("%s: %w", op, ErrInvalidEventTypeName)
	}

	if !isValidPrimaryPersonsMode(mode) {
		return models.EventType{}, fmt.Errorf("%s: %w", op, ErrInvalidPrimaryPersonsMode)
	}

	if mode == models.PrimaryPersonsModeFixed && count < 1 {
		return models.EventType{}, fmt.Errorf("%s: %w", op, ErrInvalidPrimaryPersonsCount)
	}
	if mode == models.PrimaryPersonsModeUnlimited {
		count = 0
	}

	now := time.Now()
	eventType := models.EventType{
		ID:                  uuid.New(),
		OwnerUserID:         requestUserID,
		IsSystem:            false,
		Name:                normalizedName,
		PrimaryPersonsMode:  mode,
		PrimaryPersonsCount: count,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.storage.CreateEventType(ctx, eventType); err != nil {
		if errors.Is(err, storage.ErrEventTypeAlreadyExists) {
			return models.EventType{}, fmt.Errorf("%s: %w", op, ErrEventTypeAlreadyExists)
		}
		return models.EventType{}, fmt.Errorf("%s: %w", op, err)
	}

	return eventType, nil
}

func (s *Service) DeleteEventType(ctx context.Context, requestUserID int, eventTypeID string) error {
	const op = "service.events.DeleteEventType"

	if requestUserID <= 0 {
		return fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	parsedEventTypeID, err := uuid.Parse(eventTypeID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidEventTypeID)
	}

	eventType, err := s.storage.GetEventType(ctx, parsedEventTypeID)
	if err != nil {
		if errors.Is(err, storage.ErrEventTypeNotFound) {
			return fmt.Errorf("%s: %w", op, ErrEventTypeNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if eventType.IsSystem {
		return fmt.Errorf("%s: %w", op, ErrCannotDeleteSystemEventType)
	}

	if eventType.OwnerUserID != requestUserID {
		return fmt.Errorf("%s: %w", op, ErrForbidden)
	}

	hasEvents, err := s.storage.HasEventsByType(ctx, parsedEventTypeID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if hasEvents {
		return fmt.Errorf("%s: %w", op, ErrEventTypeInUse)
	}

	if err := s.storage.DeleteEventType(ctx, parsedEventTypeID); err != nil {
		if errors.Is(err, storage.ErrEventTypeNotFound) {
			return fmt.Errorf("%s: %w", op, ErrEventTypeNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) GetEventType(ctx context.Context, requestUserID int, eventTypeID string) (models.EventType, error) {
	const op = "service.events.GetEventType"

	if requestUserID <= 0 {
		return models.EventType{}, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	parsedEventTypeID, err := uuid.Parse(eventTypeID)
	if err != nil {
		return models.EventType{}, fmt.Errorf("%s: %w", op, ErrInvalidEventTypeID)
	}

	eventType, err := s.loadAndAuthorizeEventType(ctx, op, requestUserID, parsedEventTypeID)
	if err != nil {
		return models.EventType{}, err
	}

	return eventType, nil
}

func (s *Service) ListEventTypes(ctx context.Context, requestUserID int) ([]models.EventType, error) {
	const op = "service.events.ListEventTypes"

	if requestUserID <= 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	eventTypes, err := s.storage.ListEventTypesForUser(ctx, requestUserID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return eventTypes, nil
}

func (s *Service) CreateEvent(
	ctx context.Context,
	requestUserID int,
	treeID string,
	eventTypeID string,
	primaryPersonIDs []string,
	additionalPersonIDs []string,
	dateISO string,
	datePrecision models.EventDatePrecision,
	dateBound models.EventDateBound,
) (models.Event, error) {
	const op = "service.events.CreateEvent"

	parsedTreeID, parsedEventTypeID, parsedDate, err := s.parseAndValidateBaseInput(
		op,
		requestUserID,
		treeID,
		eventTypeID,
		dateISO,
		datePrecision,
		dateBound,
	)
	if err != nil {
		return models.Event{}, err
	}

	eventType, err := s.loadAndAuthorizeEventType(ctx, op, requestUserID, parsedEventTypeID)
	if err != nil {
		return models.Event{}, err
	}

	primaryParsed, additionalParsed, participantIDs, err := validateAndParseParticipants(primaryPersonIDs, additionalPersonIDs, eventType)
	if err != nil {
		return models.Event{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.familyTree.ValidatePersonsInTree(ctx, requestUserID, parsedTreeID.String(), participantIDs); err != nil {
		return models.Event{}, fmt.Errorf("%s: %w", op, err)
	}

	now := time.Now()
	event := models.Event{
		ID:                  uuid.New(),
		TreeID:              parsedTreeID,
		EventTypeID:         parsedEventTypeID,
		PrimaryPersonIDs:    primaryParsed,
		AdditionalPersonIDs: additionalParsed,
		DateValue:           parsedDate,
		DatePrecision:       datePrecision,
		DateBound:           dateBound,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.storage.CreateEvent(ctx, event); err != nil {
		return models.Event{}, fmt.Errorf("%s: %w", op, err)
	}

	return event, nil
}

func (s *Service) UpdateEvent(
	ctx context.Context,
	requestUserID int,
	eventID string,
	eventTypeID string,
	primaryPersonIDs []string,
	additionalPersonIDs []string,
	dateISO string,
	datePrecision models.EventDatePrecision,
	dateBound models.EventDateBound,
) (models.Event, error) {
	const op = "service.events.UpdateEvent"

	if requestUserID <= 0 {
		return models.Event{}, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	parsedEventID, err := uuid.Parse(eventID)
	if err != nil {
		return models.Event{}, fmt.Errorf("%s: %w", op, ErrInvalidEventID)
	}

	existing, err := s.storage.GetEvent(ctx, parsedEventID)
	if err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			return models.Event{}, fmt.Errorf("%s: %w", op, ErrEventNotFound)
		}
		return models.Event{}, fmt.Errorf("%s: %w", op, err)
	}

	parsedEventTypeID, err := uuid.Parse(eventTypeID)
	if err != nil {
		return models.Event{}, fmt.Errorf("%s: %w", op, ErrInvalidEventTypeID)
	}

	if !isValidEventDatePrecision(datePrecision) {
		return models.Event{}, fmt.Errorf("%s: %w", op, ErrInvalidEventDatePrecision)
	}
	if !isValidEventDateBound(dateBound) {
		return models.Event{}, fmt.Errorf("%s: %w", op, ErrInvalidEventDateBound)
	}

	parsedDate, err := time.Parse("2006-01-02", dateISO)
	if err != nil {
		return models.Event{}, fmt.Errorf("%s: %w", op, ErrInvalidEventDate)
	}

	eventType, err := s.loadAndAuthorizeEventType(ctx, op, requestUserID, parsedEventTypeID)
	if err != nil {
		return models.Event{}, err
	}

	primaryParsed, additionalParsed, participantIDs, err := validateAndParseParticipants(primaryPersonIDs, additionalPersonIDs, eventType)
	if err != nil {
		return models.Event{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.familyTree.ValidatePersonsInTree(ctx, requestUserID, existing.TreeID.String(), participantIDs); err != nil {
		return models.Event{}, fmt.Errorf("%s: %w", op, err)
	}

	existing.EventTypeID = parsedEventTypeID
	existing.PrimaryPersonIDs = primaryParsed
	existing.AdditionalPersonIDs = additionalParsed
	existing.DateValue = parsedDate
	existing.DatePrecision = datePrecision
	existing.DateBound = dateBound
	existing.UpdatedAt = time.Now()

	if err := s.storage.UpdateEvent(ctx, existing); err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			return models.Event{}, fmt.Errorf("%s: %w", op, ErrEventNotFound)
		}
		return models.Event{}, fmt.Errorf("%s: %w", op, err)
	}

	return existing, nil
}

func (s *Service) DeleteEvent(ctx context.Context, requestUserID int, eventID string) error {
	const op = "service.events.DeleteEvent"

	if requestUserID <= 0 {
		return fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	parsedEventID, err := uuid.Parse(eventID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidEventID)
	}

	event, err := s.storage.GetEvent(ctx, parsedEventID)
	if err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			return fmt.Errorf("%s: %w", op, ErrEventNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.familyTree.ValidatePersonsInTree(ctx, requestUserID, event.TreeID.String(), nil); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.storage.DeleteEvent(ctx, parsedEventID); err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			return fmt.Errorf("%s: %w", op, ErrEventNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) GetEvent(ctx context.Context, requestUserID int, eventID string) (models.Event, error) {
	const op = "service.events.GetEvent"

	if requestUserID <= 0 {
		return models.Event{}, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	parsedEventID, err := uuid.Parse(eventID)
	if err != nil {
		return models.Event{}, fmt.Errorf("%s: %w", op, ErrInvalidEventID)
	}

	event, err := s.storage.GetEvent(ctx, parsedEventID)
	if err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			return models.Event{}, fmt.Errorf("%s: %w", op, ErrEventNotFound)
		}
		return models.Event{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.familyTree.ValidatePersonsInTree(ctx, requestUserID, event.TreeID.String(), nil); err != nil {
		return models.Event{}, fmt.Errorf("%s: %w", op, err)
	}

	return event, nil
}

func (s *Service) ListEventsByTree(ctx context.Context, requestUserID int, treeID string) ([]models.Event, error) {
	const op = "service.events.ListEventsByTree"

	if requestUserID <= 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	if err := s.familyTree.ValidatePersonsInTree(ctx, requestUserID, parsedTreeID.String(), nil); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	events, err := s.storage.ListEventsByTree(ctx, parsedTreeID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return events, nil
}

func (s *Service) parseAndValidateBaseInput(
	op string,
	requestUserID int,
	treeID string,
	eventTypeID string,
	dateISO string,
	datePrecision models.EventDatePrecision,
	dateBound models.EventDateBound,
) (uuid.UUID, uuid.UUID, time.Time, error) {
	if requestUserID <= 0 {
		return uuid.Nil, uuid.Nil, time.Time{}, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return uuid.Nil, uuid.Nil, time.Time{}, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	parsedEventTypeID, err := uuid.Parse(eventTypeID)
	if err != nil {
		return uuid.Nil, uuid.Nil, time.Time{}, fmt.Errorf("%s: %w", op, ErrInvalidEventTypeID)
	}

	if !isValidEventDatePrecision(datePrecision) {
		return uuid.Nil, uuid.Nil, time.Time{}, fmt.Errorf("%s: %w", op, ErrInvalidEventDatePrecision)
	}
	if !isValidEventDateBound(dateBound) {
		return uuid.Nil, uuid.Nil, time.Time{}, fmt.Errorf("%s: %w", op, ErrInvalidEventDateBound)
	}

	parsedDate, err := time.Parse("2006-01-02", dateISO)
	if err != nil {
		return uuid.Nil, uuid.Nil, time.Time{}, fmt.Errorf("%s: %w", op, ErrInvalidEventDate)
	}

	return parsedTreeID, parsedEventTypeID, parsedDate, nil
}

func (s *Service) loadAndAuthorizeEventType(ctx context.Context, op string, requestUserID int, eventTypeID uuid.UUID) (models.EventType, error) {
	eventType, err := s.storage.GetEventType(ctx, eventTypeID)
	if err != nil {
		if errors.Is(err, storage.ErrEventTypeNotFound) {
			return models.EventType{}, fmt.Errorf("%s: %w", op, ErrEventTypeNotFound)
		}
		return models.EventType{}, fmt.Errorf("%s: %w", op, err)
	}

	if !eventType.IsSystem && eventType.OwnerUserID != requestUserID {
		return models.EventType{}, fmt.Errorf("%s: %w", op, ErrForbidden)
	}

	return eventType, nil
}

func validateAndParseParticipants(primary []string, additional []string, eventType models.EventType) ([]uuid.UUID, []uuid.UUID, []string, error) {
	if !isValidPrimaryPersonsMode(eventType.PrimaryPersonsMode) {
		return nil, nil, nil, ErrInvalidPrimaryPersonsMode
	}

	if len(primary) == 0 {
		return nil, nil, nil, ErrInvalidPrimaryPersons
	}

	if eventType.PrimaryPersonsMode == models.PrimaryPersonsModeFixed && len(primary) != eventType.PrimaryPersonsCount {
		return nil, nil, nil, ErrInvalidPrimaryPersons
	}

	parsedPrimary := make([]uuid.UUID, 0, len(primary))
	parsedAdditional := make([]uuid.UUID, 0, len(additional))
	forValidation := make([]string, 0, len(primary)+len(additional))

	seenPrimary := make(map[uuid.UUID]struct{}, len(primary))
	for _, raw := range primary {
		parsedID, err := uuid.Parse(raw)
		if err != nil {
			return nil, nil, nil, ErrInvalidPrimaryPersons
		}
		if _, ok := seenPrimary[parsedID]; ok {
			return nil, nil, nil, ErrDuplicatePersonInParticipants
		}
		seenPrimary[parsedID] = struct{}{}
		parsedPrimary = append(parsedPrimary, parsedID)
		forValidation = append(forValidation, parsedID.String())
	}

	seenAdditional := make(map[uuid.UUID]struct{}, len(additional))
	for _, raw := range additional {
		parsedID, err := uuid.Parse(raw)
		if err != nil {
			return nil, nil, nil, ErrInvalidPrimaryPersons
		}
		if _, ok := seenPrimary[parsedID]; ok {
			return nil, nil, nil, ErrParticipantListsOverlap
		}
		if _, ok := seenAdditional[parsedID]; ok {
			return nil, nil, nil, ErrDuplicatePersonInParticipants
		}
		seenAdditional[parsedID] = struct{}{}
		parsedAdditional = append(parsedAdditional, parsedID)
		forValidation = append(forValidation, parsedID.String())
	}

	return parsedPrimary, parsedAdditional, forValidation, nil
}

func isValidPrimaryPersonsMode(mode models.PrimaryPersonsMode) bool {
	return mode == models.PrimaryPersonsModeFixed || mode == models.PrimaryPersonsModeUnlimited
}

func isValidEventDatePrecision(precision models.EventDatePrecision) bool {
	return precision == models.EventDatePrecisionDay ||
		precision == models.EventDatePrecisionMonth ||
		precision == models.EventDatePrecisionYear
}

func isValidEventDateBound(bound models.EventDateBound) bool {
	return bound == models.EventDateBoundExact ||
		bound == models.EventDateBoundNotBefore ||
		bound == models.EventDateBoundNotAfter
}
