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
	GetTreeAvatar(ctx context.Context, treeID uuid.UUID) (models.Photo, error)
	GetPersonAvatar(ctx context.Context, personID uuid.UUID) (models.Photo, error)
	UnsetTreeAvatar(ctx context.Context, treeID uuid.UUID) error
	UnsetPersonAvatar(ctx context.Context, personID uuid.UUID) error
	ListPersonPhotos(ctx context.Context, personID uuid.UUID) ([]models.Photo, error)
	ListEventPhotos(ctx context.Context, eventID uuid.UUID) ([]models.Photo, error)
	DeletePhotoByID(ctx context.Context, photoID uuid.UUID) (models.Photo, error)
	Close()
}

type PublicMetadataStorage interface {
	ListPublicPersonPhotos(ctx context.Context, publicPersonID uuid.UUID) ([]models.Photo, error)
	UnsetPublicPersonAvatar(ctx context.Context, publicPersonID uuid.UUID) error
	DeletePublicPersonMedia(ctx context.Context, publicPersonID uuid.UUID) ([]models.Photo, error)
}

type ObjectStorage interface {
	PutObject(ctx context.Context, key string, content []byte, mimeType string) error
	GetObject(ctx context.Context, key string) ([]byte, error)
	DeleteObject(ctx context.Context, key string) error
}
