package storage

import (
	"context"
	"errors"

	"github.com/Bad-Utya/myforebears-backend/services/events/internal/domain/models"
	"github.com/google/uuid"
)

var (
	ErrEventTypeNotFound      = errors.New("event type not found")
	ErrEventNotFound          = errors.New("event not found")
	ErrEventTypeAlreadyExists = errors.New("event type already exists")
)

type Storage interface {
	CreateEventType(ctx context.Context, eventType models.EventType) error
	GetEventType(ctx context.Context, eventTypeID uuid.UUID) (models.EventType, error)
	ListEventTypesForUser(ctx context.Context, requestUserID int) ([]models.EventType, error)
	DeleteEventType(ctx context.Context, eventTypeID uuid.UUID) error
	HasEventsByType(ctx context.Context, eventTypeID uuid.UUID) (bool, error)

	CreateEvent(ctx context.Context, event models.Event) error
	GetEvent(ctx context.Context, eventID uuid.UUID) (models.Event, error)
	ListEventsByTree(ctx context.Context, treeID uuid.UUID) ([]models.Event, error)
	UpdateEvent(ctx context.Context, event models.Event) error
	DeleteEvent(ctx context.Context, eventID uuid.UUID) error

	Close()
}
