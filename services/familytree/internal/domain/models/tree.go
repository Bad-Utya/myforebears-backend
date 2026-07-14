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
	Description        *string
	RootPersonID       uuid.UUID
	Tags               []Tag
	SimilarityScore    float64
}

type Tag struct {
	Code        string
	Name        string
	Description string
}
