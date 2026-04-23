package storage

import (
	"context"
	"errors"

	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/domain/models"
	"github.com/google/uuid"
)

var (
	ErrPhotoNotFound = errors.New("photo not found")
)

type MetadataStorage interface {
	CreatePhoto(ctx context.Context, photo models.Photo) error
	GetPhotoByID(ctx context.Context, photoID uuid.UUID) (models.Photo, error)
	GetUserAvatar(ctx context.Context, ownerUserID int) (models.Photo, error)
	GetPersonAvatar(ctx context.Context, personID uuid.UUID) (models.Photo, error)
	UnsetPersonAvatar(ctx context.Context, personID uuid.UUID) error
	ListPersonPhotos(ctx context.Context, personID uuid.UUID) ([]models.Photo, error)
	ListEventPhotos(ctx context.Context, eventID uuid.UUID) ([]models.Photo, error)
	DeletePhotoByID(ctx context.Context, photoID uuid.UUID) (models.Photo, error)
	Close()
}

type ObjectStorage interface {
	PutObject(ctx context.Context, key string, content []byte, mimeType string) error
	GetObject(ctx context.Context, key string) ([]byte, error)
	DeleteObject(ctx context.Context, key string) error
}
