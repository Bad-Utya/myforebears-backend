package models

import (
	"time"

	"github.com/google/uuid"
)

type Tree struct {
	ID                 uuid.UUID
	CreatorID          int
	CreatedAt          time.Time
	IsViewRestricted   bool
	IsPublicOnMainPage bool
	Name               string
}
