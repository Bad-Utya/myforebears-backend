package storage

import (
	"context"
	"errors"

	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/domain/models"
	"github.com/google/uuid"
)

var (
	ErrVisualisationNotFound = errors.New("visualisation not found")
)

type MetadataStorage interface {
	CreateVisualisation(ctx context.Context, vis models.Visualisation) error
	GetVisualisationByID(ctx context.Context, visualisationID uuid.UUID) (models.Visualisation, error)
	ListTreeVisualisations(ctx context.Context, treeID uuid.UUID) ([]models.Visualisation, error)
	DeleteVisualisationByID(ctx context.Context, visualisationID uuid.UUID) (models.Visualisation, error)
	SetVisualisationProcessing(ctx context.Context, visualisationID uuid.UUID) error
	SetVisualisationFailed(ctx context.Context, visualisationID uuid.UUID, errorMessage string) error
	SetVisualisationReady(ctx context.Context, visualisationID uuid.UUID, sizeBytes int64) error
	Close()
}

type ObjectStorage interface {
	PutObject(ctx context.Context, key string, content []byte, mimeType string) error
	GetObject(ctx context.Context, key string) ([]byte, error)
	DeleteObject(ctx context.Context, key string) error
}
