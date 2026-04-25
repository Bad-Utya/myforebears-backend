package models

import (
	"time"

	"github.com/google/uuid"
)

type VisualisationType string

const (
	VisualisationTypeAncestors               VisualisationType = "ancestors"
	VisualisationTypeDescendants             VisualisationType = "descendants"
	VisualisationTypeAncestorsAndDescendants VisualisationType = "ancestors_and_descendants"
	VisualisationTypeFull                    VisualisationType = "full"
)

type VisualisationStatus string

const (
	VisualisationStatusPending    VisualisationStatus = "pending"
	VisualisationStatusProcessing VisualisationStatus = "processing"
	VisualisationStatusReady      VisualisationStatus = "ready"
	VisualisationStatusFailed     VisualisationStatus = "failed"
)

type Visualisation struct {
	ID                uuid.UUID
	OwnerUserID       int
	TreeID            uuid.UUID
	RootPersonID      uuid.UUID
	IncludedPersonIDs []uuid.UUID
	Type              VisualisationType
	Status            VisualisationStatus
	FileName          string
	MIMEType          string
	SizeBytes         int64
	ObjectKey         string
	ErrorMessage      string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CompletedAt       *time.Time
}
