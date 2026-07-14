package models

import (
	"time"

	"github.com/google/uuid"
)

type Photo struct {
	ID             uuid.UUID
	OwnerUserID    int
	TreeID         *uuid.UUID
	PersonID       *uuid.UUID
	EventID        *uuid.UUID
	PublicPersonID *uuid.UUID
	PublicEventID  *uuid.UUID
	IsUserAvatar   bool
	IsTreeAvatar   bool
	IsPersonAvatar bool
	FileName       string
	MIMEType       string
	SizeBytes      int64
	ObjectKey      string
	CreatedAt      time.Time
}
