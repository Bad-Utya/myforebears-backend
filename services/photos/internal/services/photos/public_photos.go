package photos

import (
	"context"
	"fmt"
	"time"

	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/domain/models"
	"github.com/google/uuid"
)

type EventPhotoMapping struct{ SourceEventID, TargetEventID uuid.UUID }

func (s *Service) CopyPersonMediaToPublic(ctx context.Context, userID int, treeID, personID, publicPersonID string, mappings []EventPhotoMapping) ([]models.Photo, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}
	owner, err := s.familyTree.GetTreeCreatorID(ctx, treeID)
	if err != nil {
		return nil, err
	}
	if owner != userID {
		return nil, ErrForbidden
	}
	if err := s.familyTree.GetPerson(ctx, treeID, personID); err != nil {
		return nil, err
	}
	personUUID, err := uuid.Parse(personID)
	if err != nil {
		return nil, ErrInvalidPersonID
	}
	publicUUID, err := uuid.Parse(publicPersonID)
	if err != nil {
		return nil, ErrInvalidPersonID
	}
	sources, err := s.meta.ListPersonPhotos(ctx, personUUID)
	if err != nil {
		return nil, err
	}
	for _, mapping := range mappings {
		photos, err := s.meta.ListEventPhotos(ctx, mapping.SourceEventID)
		if err != nil {
			return nil, err
		}
		for i := range photos {
			photos[i].PublicEventID = &mapping.TargetEventID
		}
		sources = append(sources, photos...)
	}
	created, err := s.copyPhotos(ctx, sources, func(source models.Photo) models.Photo {
		source.ID = uuid.New()
		source.OwnerUserID = userID
		source.TreeID = nil
		source.PersonID = nil
		source.EventID = nil
		source.PublicPersonID = &publicUUID
		source.ObjectKey = buildObjectKey("public-persons", publicUUID.String(), "media", source.FileName)
		source.CreatedAt = time.Now()
		return source
	})
	if err != nil {
		return nil, err
	}
	for _, photo := range created {
		if photo.IsPersonAvatar {
			if err := s.publicTree.SetPublicPersonAvatarPhoto(ctx, userID, publicPersonID, photo.ID.String()); err != nil {
				return nil, err
			}
			break
		}
	}
	return created, nil
}

func (s *Service) CopyPublicPersonMediaToTree(ctx context.Context, userID int, publicPersonID, treeID, personID string, mappings []EventPhotoMapping) ([]models.Photo, error) {
	owner, err := s.familyTree.GetTreeCreatorID(ctx, treeID)
	if err != nil {
		return nil, err
	}
	if owner != userID {
		return nil, ErrForbidden
	}
	publicUUID, err := uuid.Parse(publicPersonID)
	if err != nil {
		return nil, ErrInvalidPersonID
	}
	treeUUID, err := uuid.Parse(treeID)
	if err != nil {
		return nil, ErrInvalidTreeID
	}
	personUUID, err := uuid.Parse(personID)
	if err != nil {
		return nil, ErrInvalidPersonID
	}
	sources, err := s.publicMeta.ListPublicPersonPhotos(ctx, publicUUID)
	if err != nil {
		return nil, err
	}
	eventMap := make(map[uuid.UUID]uuid.UUID, len(mappings))
	for _, m := range mappings {
		eventMap[m.SourceEventID] = m.TargetEventID
	}
	filtered := make([]models.Photo, 0, len(sources))
	for _, source := range sources {
		if source.PublicEventID != nil {
			target, ok := eventMap[*source.PublicEventID]
			if !ok {
				continue
			}
			source.EventID = &target
			source.PersonID = nil
		} else {
			source.PersonID = &personUUID
		}
		filtered = append(filtered, source)
	}
	created, err := s.copyPhotos(ctx, filtered, func(source models.Photo) models.Photo {
		source.ID = uuid.New()
		source.OwnerUserID = userID
		source.TreeID = &treeUUID
		source.PublicPersonID = nil
		source.PublicEventID = nil
		source.ObjectKey = buildObjectKey("persons", personUUID.String(), "imported", source.FileName)
		source.CreatedAt = time.Now()
		return source
	})
	if err != nil {
		return nil, err
	}
	for _, photo := range created {
		if photo.IsPersonAvatar {
			if err := s.familyTree.UpdatePersonAvatarPhoto(ctx, personID, photo.ID.String()); err != nil {
				return nil, err
			}
			break
		}
	}
	return created, nil
}

func (s *Service) UploadPublicPersonPhoto(ctx context.Context, userID int, publicPersonID, fileName, mimeType string, content []byte, isAvatar bool) (models.Photo, error) {
	if err := validateFileInput(fileName, mimeType, content); err != nil {
		return models.Photo{}, err
	}
	owner, err := s.publicTree.GetPublicPersonOwnerID(ctx, publicPersonID)
	if err != nil {
		return models.Photo{}, err
	}
	if owner != userID {
		return models.Photo{}, ErrForbidden
	}
	id, err := uuid.Parse(publicPersonID)
	if err != nil {
		return models.Photo{}, ErrInvalidPersonID
	}
	if isAvatar {
		if err := s.publicMeta.UnsetPublicPersonAvatar(ctx, id); err != nil {
			return models.Photo{}, err
		}
	}
	photo := models.Photo{ID: uuid.New(), OwnerUserID: userID, PublicPersonID: &id, IsPersonAvatar: isAvatar, FileName: normalizeFileName(fileName), MIMEType: mimeType, SizeBytes: int64(len(content)), ObjectKey: buildObjectKey("public-persons", id.String(), "gallery", fileName), CreatedAt: time.Now()}
	if err := s.objects.PutObject(ctx, photo.ObjectKey, content, mimeType); err != nil {
		return models.Photo{}, err
	}
	if err := s.meta.CreatePhoto(ctx, photo); err != nil {
		_ = s.objects.DeleteObject(ctx, photo.ObjectKey)
		return models.Photo{}, err
	}
	if isAvatar {
		if err := s.publicTree.SetPublicPersonAvatarPhoto(ctx, userID, publicPersonID, photo.ID.String()); err != nil {
			return models.Photo{}, err
		}
	}
	return photo, nil
}

func (s *Service) ListPublicPersonPhotos(ctx context.Context, publicPersonID string) ([]models.Photo, error) {
	id, err := uuid.Parse(publicPersonID)
	if err != nil {
		return nil, ErrInvalidPersonID
	}
	return s.publicMeta.ListPublicPersonPhotos(ctx, id)
}

func (s *Service) GetPublicPersonPhoto(ctx context.Context, publicPersonID, photoID string) (models.Photo, []byte, error) {
	pid, err := uuid.Parse(publicPersonID)
	if err != nil {
		return models.Photo{}, nil, ErrInvalidPersonID
	}
	id, err := uuid.Parse(photoID)
	if err != nil {
		return models.Photo{}, nil, ErrInvalidPhotoID
	}
	p, err := s.meta.GetPhotoByID(ctx, id)
	if err != nil {
		return models.Photo{}, nil, ErrPhotoNotFound
	}
	if p.PublicPersonID == nil || *p.PublicPersonID != pid {
		return models.Photo{}, nil, ErrForbidden
	}
	data, err := s.objects.GetObject(ctx, p.ObjectKey)
	return p, data, err
}

func (s *Service) DeletePublicPersonPhoto(ctx context.Context, userID int, publicPersonID, photoID string) error {
	owner, err := s.publicTree.GetPublicPersonOwnerID(ctx, publicPersonID)
	if err != nil {
		return err
	}
	if owner != userID {
		return ErrForbidden
	}
	pid, err := uuid.Parse(publicPersonID)
	if err != nil {
		return ErrInvalidPersonID
	}
	id, err := uuid.Parse(photoID)
	if err != nil {
		return ErrInvalidPhotoID
	}
	p, err := s.meta.GetPhotoByID(ctx, id)
	if err != nil {
		return ErrPhotoNotFound
	}
	if p.PublicPersonID == nil || *p.PublicPersonID != pid {
		return ErrForbidden
	}
	deleted, err := s.meta.DeletePhotoByID(ctx, id)
	if err != nil {
		return err
	}
	if deleted.IsPersonAvatar {
		_ = s.publicTree.SetPublicPersonAvatarPhoto(ctx, userID, publicPersonID, "")
	}
	if err := s.objects.DeleteObject(ctx, deleted.ObjectKey); err != nil {
		s.log.Error("failed to delete public photo object", "error", err)
	}
	return nil
}

func (s *Service) DeletePublicPersonMedia(ctx context.Context, userID int, publicPersonID string) error {
	owner, err := s.publicTree.GetPublicPersonOwnerID(ctx, publicPersonID)
	if err != nil {
		return err
	}
	if owner != userID {
		return ErrForbidden
	}
	id, err := uuid.Parse(publicPersonID)
	if err != nil {
		return ErrInvalidPersonID
	}
	photos, err := s.publicMeta.DeletePublicPersonMedia(ctx, id)
	if err != nil {
		return err
	}
	for _, p := range photos {
		if err := s.objects.DeleteObject(ctx, p.ObjectKey); err != nil {
			s.log.Error("failed to delete public media object", "error", err)
		}
	}
	return nil
}

func (s *Service) copyPhotos(ctx context.Context, sources []models.Photo, transform func(models.Photo) models.Photo) ([]models.Photo, error) {
	created := make([]models.Photo, 0, len(sources))
	for _, source := range sources {
		data, err := s.objects.GetObject(ctx, source.ObjectKey)
		if err != nil {
			s.rollbackCopiedPhotos(ctx, created)
			return nil, fmt.Errorf("copy photo: %w", err)
		}
		target := transform(source)
		if err := s.objects.PutObject(ctx, target.ObjectKey, data, target.MIMEType); err != nil {
			s.rollbackCopiedPhotos(ctx, created)
			return nil, err
		}
		if err := s.meta.CreatePhoto(ctx, target); err != nil {
			_ = s.objects.DeleteObject(ctx, target.ObjectKey)
			s.rollbackCopiedPhotos(ctx, created)
			return nil, err
		}
		created = append(created, target)
	}
	return created, nil
}
func (s *Service) rollbackCopiedPhotos(ctx context.Context, photos []models.Photo) {
	for _, p := range photos {
		_, _ = s.meta.DeletePhotoByID(ctx, p.ID)
		_ = s.objects.DeleteObject(ctx, p.ObjectKey)
	}
}
