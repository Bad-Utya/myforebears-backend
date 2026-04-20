package models

import (
	"time"

	"github.com/google/uuid"
)

type PrimaryPersonsMode string

const (
	PrimaryPersonsModeFixed     PrimaryPersonsMode = "FIXED"
	PrimaryPersonsModeUnlimited PrimaryPersonsMode = "UNLIMITED"
)

type EventDatePrecision string

const (
	EventDatePrecisionDay   EventDatePrecision = "DAY"
	EventDatePrecisionMonth EventDatePrecision = "MONTH"
	EventDatePrecisionYear  EventDatePrecision = "YEAR"
)

type EventDateBound string

const (
	EventDateBoundExact     EventDateBound = "EXACT"
	EventDateBoundNotBefore EventDateBound = "NOT_BEFORE"
	EventDateBoundNotAfter  EventDateBound = "NOT_AFTER"
)

type EventType struct {
	ID                  uuid.UUID
	OwnerUserID         int
	IsSystem            bool
	Name                string
	PrimaryPersonsMode  PrimaryPersonsMode
	PrimaryPersonsCount int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Event struct {
	ID                  uuid.UUID
	TreeID              uuid.UUID
	EventTypeID         uuid.UUID
	PrimaryPersonIDs    []uuid.UUID
	AdditionalPersonIDs []uuid.UUID
	DateValue           time.Time
	DatePrecision       EventDatePrecision
	DateBound           EventDateBound
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
