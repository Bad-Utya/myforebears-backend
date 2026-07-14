package models

import (
	"time"

	"github.com/google/uuid"
)

type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
)

type Person struct {
	ID            uuid.UUID
	TreeID        uuid.UUID
	FirstName     string
	LastName      string
	Patronymic    string
	Gender        Gender
	Biography     string
	AvatarPhotoID *uuid.UUID
}

type PublicPersonEvent struct {
	ID             uuid.UUID
	PublicPersonID uuid.UUID
	SourceEventID  *uuid.UUID
	EventTypeID    *uuid.UUID
	EventTypeName  string
	DateISO        string
	DatePrecision  string
	DateBound      string
	DateUnknown    bool
}

type PublicPerson struct {
	ID              uuid.UUID
	OwnerUserID     int
	FirstName       string
	LastName        string
	Patronymic      string
	Gender          Gender
	Biography       string
	AvatarPhotoID   *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Events          []PublicPersonEvent
	Tags            []Tag
	SimilarityScore float64
}
