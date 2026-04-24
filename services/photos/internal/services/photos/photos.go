package photos

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/storage"
	"github.com/google/uuid"
)

const maxPhotoSizeBytes = 15 * 1024 * 1024

var (
	ErrInvalidUserID   = errors.New("invalid user id")
	ErrInvalidPhotoID  = errors.New("invalid photo id")
	ErrInvalidTreeID   = errors.New("invalid tree id")
	ErrInvalidPersonID = errors.New("invalid person id")
	ErrInvalidEventID  = errors.New("invalid event id")
	ErrInvalidFileName = errors.New("file name is required")
	ErrInvalidMIMEType = errors.New("mime type must start with image/")
	ErrEmptyContent    = errors.New("photo content is empty")
	ErrTooLarge        = errors.New("photo is too large")
	ErrPhotoNotFound   = errors.New("photo not found")
	ErrForbidden       = errors.New("forbidden")
)

type FamilyTreeClient interface {
	GetPerson(ctx context.Context, treeID string, personID string) error
	UpdatePersonAvatarPhoto(ctx context.Context, personID string, avatarPhotoID string) error
}

type EventsClient interface {
	GetEventTreeID(ctx context.Context, requestUserID int, eventID string) (string, error)
}

type Service struct {
	log        *slog.Logger
	meta       storage.MetadataStorage
	objects    storage.ObjectStorage
	familyTree FamilyTreeClient
	events     EventsClient
}

func New(log *slog.Logger, meta storage.MetadataStorage, objects storage.ObjectStorage, familyTree FamilyTreeClient, events EventsClient) *Service {
	return &Service{log: log, meta: meta, objects: objects, familyTree: familyTree, events: events}
}

func (s *Service) UploadUserAvatar(ctx context.Context, requestUserID int, fileName string, mimeType string, content []byte) (models.Photo, error) {
	const op = "service.photos.UploadUserAvatar"
	log := s.log.With(slog.String("op", op), slog.Int("request_user_id", requestUserID))

	if err := validateFileInput(requestUserID, fileName, mimeType, content); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	oldAvatar, err := s.meta.GetUserAvatar(ctx, requestUserID)
	if err == nil {
		if err := s.objects.DeleteObject(ctx, oldAvatar.ObjectKey); err != nil {
			log.Error("failed to delete previous avatar object", slog.String("error", err.Error()))
			return models.Photo{}, fmt.Errorf("%s: %w", op, err)
		}
		if _, err := s.meta.DeletePhotoByID(ctx, oldAvatar.ID); err != nil {
			log.Error("failed to delete previous avatar metadata", slog.String("error", err.Error()))
			return models.Photo{}, fmt.Errorf("%s: %w", op, err)
		}
	} else if !errors.Is(err, storage.ErrPhotoNotFound) {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	photo := models.Photo{
		ID:             uuid.New(),
		OwnerUserID:    requestUserID,
		IsUserAvatar:   true,
		IsPersonAvatar: false,
		FileName:       normalizeFileName(fileName),
		MIMEType:       mimeType,
		SizeBytes:      int64(len(content)),
		ObjectKey:      buildObjectKey("users", fmt.Sprintf("%d", requestUserID), "avatar", fileName),
		CreatedAt:      time.Now(),
	}

	if err := s.objects.PutObject(ctx, photo.ObjectKey, content, photo.MIMEType); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}
	if err := s.meta.CreatePhoto(ctx, photo); err != nil {
		_ = s.objects.DeleteObject(ctx, photo.ObjectKey)
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	return photo, nil
}

func (s *Service) GetUserAvatar(ctx context.Context, requestUserID int) (models.Photo, []byte, error) {
	const op = "service.photos.GetUserAvatar"

	if requestUserID <= 0 {
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	photo, err := s.meta.GetUserAvatar(ctx, requestUserID)
	if err != nil {
		if errors.Is(err, storage.ErrPhotoNotFound) {
			return models.Photo{}, nil, fmt.Errorf("%s: %w", op, ErrPhotoNotFound)
		}
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	data, err := s.objects.GetObject(ctx, photo.ObjectKey)
	if err != nil {
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	return photo, data, nil
}

func (s *Service) UploadPersonAvatar(ctx context.Context, requestUserID int, personID string, fileName string, mimeType string, content []byte) (models.Photo, error) {
	const op = "service.photos.UploadPersonAvatar"

	if err := validateFileInput(requestUserID, fileName, mimeType, content); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	tree_id := "" // TODO: change

	if err := s.familyTree.GetPerson(ctx, tree_id, parsedPersonID.String()); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.meta.UnsetPersonAvatar(ctx, parsedPersonID); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	photo := models.Photo{
		ID:             uuid.New(),
		OwnerUserID:    requestUserID,
		PersonID:       &parsedPersonID,
		IsPersonAvatar: true,
		FileName:       normalizeFileName(fileName),
		MIMEType:       mimeType,
		SizeBytes:      int64(len(content)),
		ObjectKey:      buildObjectKey("persons", parsedPersonID.String(), "avatar", fileName),
		CreatedAt:      time.Now(),
	}

	if err := s.objects.PutObject(ctx, photo.ObjectKey, content, photo.MIMEType); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}
	if err := s.meta.CreatePhoto(ctx, photo); err != nil {
		_ = s.objects.DeleteObject(ctx, photo.ObjectKey)
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.familyTree.UpdatePersonAvatarPhoto(ctx, parsedPersonID.String(), photo.ID.String()); err != nil {
		_, _ = s.meta.DeletePhotoByID(ctx, photo.ID)
		_ = s.objects.DeleteObject(ctx, photo.ObjectKey)
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	return photo, nil
}

func (s *Service) GetPersonAvatar(ctx context.Context, requestUserID int, personID string) (models.Photo, []byte, error) {
	const op = "service.photos.GetPersonAvatar"

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	photo, err := s.meta.GetPersonAvatar(ctx, parsedPersonID)
	if err != nil {
		if errors.Is(err, storage.ErrPhotoNotFound) {
			return models.Photo{}, nil, fmt.Errorf("%s: %w", op, ErrPhotoNotFound)
		}
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	data, err := s.objects.GetObject(ctx, photo.ObjectKey)
	if err != nil {
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	return photo, data, nil
}

func (s *Service) UploadPersonPhoto(ctx context.Context, requestUserID int, personID string, fileName string, mimeType string, content []byte) (models.Photo, error) {
	const op = "service.photos.UploadPersonPhoto"

	if err := validateFileInput(requestUserID, fileName, mimeType, content); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	tree_id := "" // TODO: change

	if err := s.familyTree.GetPerson(ctx, tree_id, parsedPersonID.String()); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	photo := models.Photo{
		ID:             uuid.New(),
		OwnerUserID:    requestUserID,
		PersonID:       &parsedPersonID,
		IsPersonAvatar: false,
		FileName:       normalizeFileName(fileName),
		MIMEType:       mimeType,
		SizeBytes:      int64(len(content)),
		ObjectKey:      buildObjectKey("persons", parsedPersonID.String(), "gallery", fileName),
		CreatedAt:      time.Now(),
	}

	if err := s.objects.PutObject(ctx, photo.ObjectKey, content, photo.MIMEType); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}
	if err := s.meta.CreatePhoto(ctx, photo); err != nil {
		_ = s.objects.DeleteObject(ctx, photo.ObjectKey)
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	return photo, nil
}

func (s *Service) ListPersonPhotos(ctx context.Context, requestUserID int, personID string) ([]models.Photo, error) {
	const op = "service.photos.ListPersonPhotos"

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tree_id := "" // TODO: change

	if err := s.familyTree.GetPerson(ctx, tree_id, parsedPersonID.String()); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	photos, err := s.meta.ListPersonPhotos(ctx, parsedPersonID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return photos, nil
}

func (s *Service) UploadEventPhoto(ctx context.Context, requestUserID int, eventID string, fileName string, mimeType string, content []byte) (models.Photo, error) {
	const op = "service.photos.UploadEventPhoto"

	if err := validateFileInput(requestUserID, fileName, mimeType, content); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	parsedEventID, err := uuid.Parse(eventID)
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, ErrInvalidEventID)
	}

	treeID, err := s.events.GetEventTreeID(ctx, requestUserID, parsedEventID.String())
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}
	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	photo := models.Photo{
		ID:          uuid.New(),
		OwnerUserID: requestUserID,
		TreeID:      &parsedTreeID,
		EventID:     &parsedEventID,
		FileName:    normalizeFileName(fileName),
		MIMEType:    mimeType,
		SizeBytes:   int64(len(content)),
		ObjectKey:   buildObjectKey("events", parsedEventID.String(), "gallery", fileName),
		CreatedAt:   time.Now(),
	}

	if err := s.objects.PutObject(ctx, photo.ObjectKey, content, photo.MIMEType); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}
	if err := s.meta.CreatePhoto(ctx, photo); err != nil {
		_ = s.objects.DeleteObject(ctx, photo.ObjectKey)
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	return photo, nil
}

func (s *Service) ListEventPhotos(ctx context.Context, requestUserID int, eventID string) ([]models.Photo, error) {
	const op = "service.photos.ListEventPhotos"

	if requestUserID <= 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}
	parsedEventID, err := uuid.Parse(eventID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidEventID)
	}

	photos, err := s.meta.ListEventPhotos(ctx, parsedEventID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return photos, nil
}

func (s *Service) GetPhotoByID(ctx context.Context, requestUserID int, photoID string) (models.Photo, []byte, error) {
	const op = "service.photos.GetPhotoByID"

	parsedPhotoID, err := validateAndParsePhotoID(requestUserID, photoID)
	if err != nil {
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	photo, err := s.meta.GetPhotoByID(ctx, parsedPhotoID)
	if err != nil {
		if errors.Is(err, storage.ErrPhotoNotFound) {
			return models.Photo{}, nil, fmt.Errorf("%s: %w", op, ErrPhotoNotFound)
		}
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	data, err := s.objects.GetObject(ctx, photo.ObjectKey)
	if err != nil {
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	return photo, data, nil
}

func (s *Service) DeletePhotoByID(ctx context.Context, requestUserID int, photoID string) error {
	const op = "service.photos.DeletePhotoByID"

	parsedPhotoID, err := validateAndParsePhotoID(requestUserID, photoID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.meta.GetPhotoByID(ctx, parsedPhotoID)
	if err != nil {
		if errors.Is(err, storage.ErrPhotoNotFound) {
			return fmt.Errorf("%s: %w", op, ErrPhotoNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	deleted, err := s.meta.DeletePhotoByID(ctx, parsedPhotoID)
	if err != nil {
		if errors.Is(err, storage.ErrPhotoNotFound) {
			return fmt.Errorf("%s: %w", op, ErrPhotoNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if deleted.IsPersonAvatar && deleted.PersonID != nil {
		if err := s.familyTree.UpdatePersonAvatarPhoto(ctx, deleted.PersonID.String(), ""); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := s.objects.DeleteObject(ctx, deleted.ObjectKey); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func validateAndParsePhotoID(requestUserID int, photoID string) (uuid.UUID, error) {
	if requestUserID <= 0 {
		return uuid.Nil, ErrInvalidUserID
	}

	parsedPhotoID, err := uuid.Parse(photoID)
	if err != nil {
		return uuid.Nil, ErrInvalidPhotoID
	}

	return parsedPhotoID, nil
}

func validateFileInput(requestUserID int, fileName string, mimeType string, content []byte) error {
	if requestUserID <= 0 {
		return ErrInvalidUserID
	}
	if strings.TrimSpace(fileName) == "" {
		return ErrInvalidFileName
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(mimeType)), "image/") {
		return ErrInvalidMIMEType
	}
	if len(content) == 0 {
		return ErrEmptyContent
	}
	if len(content) > maxPhotoSizeBytes {
		return ErrTooLarge
	}
	return nil
}

func normalizeFileName(fileName string) string {
	name := filepath.Base(strings.TrimSpace(fileName))
	if name == "." || name == "" {
		return "image"
	}
	return name
}

func buildObjectKey(scope string, ownerID string, kind string, fileName string) string {
	return fmt.Sprintf("%s/%s/%s/%s_%s", scope, ownerID, kind, uuid.NewString(), normalizeFileName(fileName))
}
