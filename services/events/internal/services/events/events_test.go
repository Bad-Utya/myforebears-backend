package events

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/events/internal/domain/models"
	"github.com/google/uuid"
)

type storageStub struct {
	t                 *testing.T
	createEventTypeFn func(ctx context.Context, eventType models.EventType) error
	getEventTypeFn    func(ctx context.Context, eventTypeID uuid.UUID) (models.EventType, error)
	listTypesFn       func(ctx context.Context, requestUserID int) ([]models.EventType, error)
	deleteTypeFn      func(ctx context.Context, eventTypeID uuid.UUID) error
	hasEventsFn       func(ctx context.Context, eventTypeID uuid.UUID) (bool, error)
	createEventFn     func(ctx context.Context, event models.Event) error
	getEventFn        func(ctx context.Context, eventID uuid.UUID) (models.Event, error)
	listEventsFn      func(ctx context.Context, treeID uuid.UUID) ([]models.Event, error)
	updateEventFn     func(ctx context.Context, event models.Event) error
	deleteEventFn     func(ctx context.Context, eventID uuid.UUID) error
}

func (s *storageStub) CreateEventType(ctx context.Context, eventType models.EventType) error {
	if s.createEventTypeFn == nil {
		s.t.Fatalf("unexpected CreateEventType call")
	}
	return s.createEventTypeFn(ctx, eventType)
}

func (s *storageStub) GetEventType(ctx context.Context, eventTypeID uuid.UUID) (models.EventType, error) {
	if s.getEventTypeFn == nil {
		s.t.Fatalf("unexpected GetEventType call")
	}
	return s.getEventTypeFn(ctx, eventTypeID)
}

func (s *storageStub) ListEventTypesForUser(ctx context.Context, requestUserID int) ([]models.EventType, error) {
	if s.listTypesFn == nil {
		s.t.Fatalf("unexpected ListEventTypesForUser call")
	}
	return s.listTypesFn(ctx, requestUserID)
}

func (s *storageStub) DeleteEventType(ctx context.Context, eventTypeID uuid.UUID) error {
	if s.deleteTypeFn == nil {
		s.t.Fatalf("unexpected DeleteEventType call")
	}
	return s.deleteTypeFn(ctx, eventTypeID)
}

func (s *storageStub) HasEventsByType(ctx context.Context, eventTypeID uuid.UUID) (bool, error) {
	if s.hasEventsFn == nil {
		s.t.Fatalf("unexpected HasEventsByType call")
	}
	return s.hasEventsFn(ctx, eventTypeID)
}

func (s *storageStub) CreateEvent(ctx context.Context, event models.Event) error {
	if s.createEventFn == nil {
		s.t.Fatalf("unexpected CreateEvent call")
	}
	return s.createEventFn(ctx, event)
}

func (s *storageStub) GetEvent(ctx context.Context, eventID uuid.UUID) (models.Event, error) {
	if s.getEventFn == nil {
		s.t.Fatalf("unexpected GetEvent call")
	}
	return s.getEventFn(ctx, eventID)
}

func (s *storageStub) ListEventsByTree(ctx context.Context, treeID uuid.UUID) ([]models.Event, error) {
	if s.listEventsFn == nil {
		s.t.Fatalf("unexpected ListEventsByTree call")
	}
	return s.listEventsFn(ctx, treeID)
}

func (s *storageStub) UpdateEvent(ctx context.Context, event models.Event) error {
	if s.updateEventFn == nil {
		s.t.Fatalf("unexpected UpdateEvent call")
	}
	return s.updateEventFn(ctx, event)
}

func (s *storageStub) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	if s.deleteEventFn == nil {
		s.t.Fatalf("unexpected DeleteEvent call")
	}
	return s.deleteEventFn(ctx, eventID)
}

func (s *storageStub) Close() {}

type familyTreeStub struct {
	t                        *testing.T
	validatePersonsInTreeFn  func(ctx context.Context, treeID string, personIDs []string) error
	updatePartnerStatusFn    func(ctx context.Context, treeID string, personID1 string, personID2 string, status familytreepb.PartnerRelationshipStatus) error
}

func (s *familyTreeStub) ValidatePersonsInTree(ctx context.Context, treeID string, personIDs []string) error {
	if s.validatePersonsInTreeFn == nil {
		s.t.Fatalf("unexpected ValidatePersonsInTree call")
	}
	return s.validatePersonsInTreeFn(ctx, treeID, personIDs)
}

func (s *familyTreeStub) UpdatePartnerRelationshipStatus(ctx context.Context, treeID string, personID1 string, personID2 string, status familytreepb.PartnerRelationshipStatus) error {
	if s.updatePartnerStatusFn == nil {
		s.t.Fatalf("unexpected UpdatePartnerRelationshipStatus call")
	}
	return s.updatePartnerStatusFn(ctx, treeID, personID1, personID2, status)
}

func TestValidateAndParseParticipantsRejectsOverlap(t *testing.T) {
	primary := []string{"11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"}
	additional := []string{"22222222-2222-2222-2222-222222222222"}

	_, _, _, err := validateAndParseParticipants(primary, additional, models.EventType{
		PrimaryPersonsMode:  models.PrimaryPersonsModeFixed,
		PrimaryPersonsCount: 2,
	})
	if !errors.Is(err, ErrParticipantListsOverlap) {
		t.Fatalf("expected ErrParticipantListsOverlap, got %v", err)
	}
}

func TestValidateAndParseParticipantsRejectsDuplicatePrimary(t *testing.T) {
	primary := []string{"11111111-1111-1111-1111-111111111111", "11111111-1111-1111-1111-111111111111"}

	_, _, _, err := validateAndParseParticipants(primary, nil, models.EventType{
		PrimaryPersonsMode:  models.PrimaryPersonsModeFixed,
		PrimaryPersonsCount: 2,
	})
	if !errors.Is(err, ErrDuplicatePersonInParticipants) {
		t.Fatalf("expected ErrDuplicatePersonInParticipants, got %v", err)
	}
}

func TestValidateAndParseParticipantsSuccess(t *testing.T) {
	primary := []string{"11111111-1111-1111-1111-111111111111"}
	additional := []string{"22222222-2222-2222-2222-222222222222"}

	primaryParsed, additionalParsed, forValidation, err := validateAndParseParticipants(primary, additional, models.EventType{
		PrimaryPersonsMode: models.PrimaryPersonsModeUnlimited,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(primaryParsed) != 1 || len(additionalParsed) != 1 {
		t.Fatalf("expected parsed participants")
	}
	if len(forValidation) != 2 {
		t.Fatalf("expected validation list with 2 ids")
	}
}

func TestCreateEventTypeTrimsAndNormalizesCount(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	var created models.EventType
	store := &storageStub{
		t: t,
		createEventTypeFn: func(ctx context.Context, eventType models.EventType) error {
			created = eventType
			return nil
		},
	}

	svc := New(log, store, &familyTreeStub{t: t})
	got, err := svc.CreateEventType(ctx, 5, "  Birthday  ", models.PrimaryPersonsModeUnlimited, 10)
	if err != nil {
		t.Fatalf("CreateEventType error: %v", err)
	}
	if got.Name != "Birthday" {
		t.Fatalf("expected trimmed name, got %q", got.Name)
	}
	if got.PrimaryPersonsCount != 0 {
		t.Fatalf("expected count 0 for unlimited, got %d", got.PrimaryPersonsCount)
	}
	if created.Name != "Birthday" {
		t.Fatalf("expected stored name to be trimmed")
	}
}

func TestLatestMaritalEventForPair(t *testing.T) {
	person1 := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	person2 := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	firstDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	secondDate := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	events := []models.Event{
		{
			ID:               uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			EventTypeID:      uuid.MustParse(marriageEventTypeID),
			PrimaryPersonIDs: []uuid.UUID{person1, person2},
			DateValue:        &firstDate,
			CreatedAt:        time.Now().Add(-2 * time.Hour),
		},
		{
			ID:               uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			EventTypeID:      uuid.MustParse(divorceEventTypeID),
			PrimaryPersonIDs: []uuid.UUID{person1, person2},
			DateValue:        &secondDate,
			CreatedAt:        time.Now().Add(-1 * time.Hour),
		},
	}

	latest := latestMaritalEventForPair(events, person1, person2)
	if latest == nil || latest.EventTypeID.String() != divorceEventTypeID {
		t.Fatalf("expected latest divorce event, got %#v", latest)
	}
}
