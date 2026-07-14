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
	GetTreeCreatorID(ctx context.Context, treeID string) (int, error)
}

type PublicFamilyTreeClient interface {
	GetPublicPersonOwnerID(ctx context.Context, publicPersonID string) (int, error)
	SetPublicPersonAvatarPhoto(ctx context.Context, requestUserID int, publicPersonID string, avatarPhotoID string) error
}

type EventsClient interface {
	IsEventFromTree(ctx context.Context, treeID string, eventID string) error
}

type Service struct {
	log        *slog.Logger
	meta       storage.MetadataStorage
	publicMeta storage.PublicMetadataStorage
	objects    storage.ObjectStorage
	familyTree FamilyTreeClient
	publicTree PublicFamilyTreeClient
	events     EventsClient
}

func New(log *slog.Logger, meta storage.MetadataStorage, objects storage.ObjectStorage, familyTree FamilyTreeClient, events EventsClient) *Service {
	publicMeta, _ := meta.(storage.PublicMetadataStorage)
	publicTree, _ := familyTree.(PublicFamilyTreeClient)
	return &Service{log: log, meta: meta, publicMeta: publicMeta, objects: objects, familyTree: familyTree, publicTree: publicTree, events: events}
}

func (s *Service) UploadUserAvatar(ctx context.Context, requestUserID int, fileName string, mimeType string, content []byte) (models.Photo, error) {
	const op = "service.photos.UploadUserAvatar"
	log := s.log.With(slog.String("op", op), slog.Int("request_user_id", requestUserID))

	if requestUserID <= 0 {
		return models.Photo{}, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	if err := validateFileInput(fileName, mimeType, content); err != nil {
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

func (s *Service) UploadTreeAvatar(ctx context.Context, treeID string, fileName string, mimeType string, content []byte) (models.Photo, error) {
	const op = "service.photos.UploadTreeAvatar"
	log := s.log.With(slog.String("op", op), slog.String("tree_id", treeID))

	if err := validateFileInput(fileName, mimeType, content); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	ownerUserID, err := s.familyTree.GetTreeCreatorID(ctx, parsedTreeID.String())
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	oldAvatar, err := s.meta.GetTreeAvatar(ctx, parsedTreeID)
	if err == nil {
		if err := s.objects.DeleteObject(ctx, oldAvatar.ObjectKey); err != nil {
			log.Error("failed to delete previous tree avatar object", slog.String("error", err.Error()))
			return models.Photo{}, fmt.Errorf("%s: %w", op, err)
		}
		if _, err := s.meta.DeletePhotoByID(ctx, oldAvatar.ID); err != nil {
			log.Error("failed to delete previous tree avatar metadata", slog.String("error", err.Error()))
			return models.Photo{}, fmt.Errorf("%s: %w", op, err)
		}
	} else if !errors.Is(err, storage.ErrPhotoNotFound) {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	photo := models.Photo{
		ID:           uuid.New(),
		OwnerUserID:  ownerUserID,
		TreeID:       &parsedTreeID,
		IsTreeAvatar: true,
		FileName:     normalizeFileName(fileName),
		MIMEType:     mimeType,
		SizeBytes:    int64(len(content)),
		ObjectKey:    buildObjectKey("trees", parsedTreeID.String(), "avatar", fileName),
		CreatedAt:    time.Now(),
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

func (s *Service) GetTreeAvatar(ctx context.Context, treeID string) (models.Photo, []byte, error) {
	const op = "service.photos.GetTreeAvatar"

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	if _, err := s.familyTree.GetTreeCreatorID(ctx, parsedTreeID.String()); err != nil {
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	photo, err := s.meta.GetTreeAvatar(ctx, parsedTreeID)
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

func (s *Service) UploadPersonAvatar(ctx context.Context, treeID string, personID string, fileName string, mimeType string, content []byte) (models.Photo, error) {
	const op = "service.photos.UploadPersonAvatar"

	if err := validateFileInput(fileName, mimeType, content); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	ownerUserID, err := s.familyTree.GetTreeCreatorID(ctx, parsedTreeID.String())
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	if err := s.familyTree.GetPerson(ctx, parsedTreeID.String(), parsedPersonID.String()); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.meta.UnsetPersonAvatar(ctx, parsedPersonID); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	photo := models.Photo{
		ID:             uuid.New(),
		OwnerUserID:    ownerUserID,
		TreeID:         &parsedTreeID,
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

func (s *Service) GetPersonAvatar(ctx context.Context, treeID string, personID string) (models.Photo, []byte, error) {
	const op = "service.photos.GetPersonAvatar"

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	if err := s.familyTree.GetPerson(ctx, parsedTreeID.String(), parsedPersonID.String()); err != nil {
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

func (s *Service) UploadPersonPhoto(ctx context.Context, treeID string, personID string, fileName string, mimeType string, content []byte) (models.Photo, error) {
	const op = "service.photos.UploadPersonPhoto"

	if err := validateFileInput(fileName, mimeType, content); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	ownerUserID, err := s.familyTree.GetTreeCreatorID(ctx, parsedTreeID.String())
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	if err := s.familyTree.GetPerson(ctx, parsedTreeID.String(), parsedPersonID.String()); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	photo := models.Photo{
		ID:             uuid.New(),
		OwnerUserID:    ownerUserID,
		TreeID:         &parsedTreeID,
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

func (s *Service) ListPersonPhotos(ctx context.Context, treeID string, personID string) ([]models.Photo, error) {
	const op = "service.photos.ListPersonPhotos"

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	if err := s.familyTree.GetPerson(ctx, parsedTreeID.String(), parsedPersonID.String()); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	photos, err := s.meta.ListPersonPhotos(ctx, parsedPersonID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return photos, nil
}

func (s *Service) UploadEventPhoto(ctx context.Context, treeID string, eventID string, fileName string, mimeType string, content []byte) (models.Photo, error) {
	const op = "service.photos.UploadEventPhoto"

	if err := validateFileInput(fileName, mimeType, content); err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	ownerUserID, err := s.familyTree.GetTreeCreatorID(ctx, parsedTreeID.String())
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	parsedEventID, err := uuid.Parse(eventID)
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, ErrInvalidEventID)
	}

	err = s.events.IsEventFromTree(ctx, parsedTreeID.String(), eventID)
	if err != nil {
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	photo := models.Photo{
		ID:          uuid.New(),
		OwnerUserID: ownerUserID,
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

func (s *Service) ListEventPhotos(ctx context.Context, treeID string, eventID string) ([]models.Photo, error) {
	const op = "service.photos.ListEventPhotos"

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	parsedEventID, err := uuid.Parse(eventID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidEventID)
	}

	if err := s.events.IsEventFromTree(ctx, parsedTreeID.String(), parsedEventID.String()); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	photos, err := s.meta.ListEventPhotos(ctx, parsedEventID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return photos, nil
}

func (s *Service) GetPhotoByID(ctx context.Context, treeID string, photoID string) (models.Photo, []byte, error) {
	const op = "service.photos.GetPhotoByID"

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	parsedPhotoID, err := validateAndParsePhotoID(photoID)
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

	if photo.IsUserAvatar || photo.TreeID == nil || *photo.TreeID != parsedTreeID {
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, ErrForbidden)
	}

	data, err := s.objects.GetObject(ctx, photo.ObjectKey)
	if err != nil {
		return models.Photo{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	return photo, data, nil
}

func (s *Service) DeletePhotoByID(ctx context.Context, treeID string, photoID string) error {
	const op = "service.photos.DeletePhotoByID"

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	parsedPhotoID, err := validateAndParsePhotoID(photoID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	photo, err := s.meta.GetPhotoByID(ctx, parsedPhotoID)
	if err != nil {
		if errors.Is(err, storage.ErrPhotoNotFound) {
			return fmt.Errorf("%s: %w", op, ErrPhotoNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if photo.IsUserAvatar || photo.TreeID == nil || *photo.TreeID != parsedTreeID {
		return fmt.Errorf("%s: %w", op, ErrForbidden)
	}

	avatarCleared := false
	if photo.IsPersonAvatar && photo.PersonID != nil {
		if err := s.familyTree.UpdatePersonAvatarPhoto(ctx, photo.PersonID.String(), ""); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		avatarCleared = true
	}

	deleted, err := s.meta.DeletePhotoByID(ctx, parsedPhotoID)
	if err != nil {
		if avatarCleared {
			if rollbackErr := s.familyTree.UpdatePersonAvatarPhoto(ctx, photo.PersonID.String(), photo.ID.String()); rollbackErr != nil {
				s.log.Error("failed to restore person avatar after metadata delete failure", slog.String("photo_id", photo.ID.String()), slog.String("error", rollbackErr.Error()))
			}
		}
		if errors.Is(err, storage.ErrPhotoNotFound) {
			return fmt.Errorf("%s: %w", op, ErrPhotoNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.objects.DeleteObject(ctx, deleted.ObjectKey); err != nil {
		// Metadata is already gone, so a failed object deletion is an orphaned
		// blob rather than a broken user-facing reference. Log it for cleanup.
		s.log.Error("failed to delete orphaned photo object", slog.String("photo_id", deleted.ID.String()), slog.String("object_key", deleted.ObjectKey), slog.String("error", err.Error()))
	}

	return nil
}

func validateAndParsePhotoID(photoID string) (uuid.UUID, error) {
	parsedPhotoID, err := uuid.Parse(photoID)
	if err != nil {
		return uuid.Nil, ErrInvalidPhotoID
	}

	return parsedPhotoID, nil
}

func validateFileInput(fileName string, mimeType string, content []byte) error {
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
