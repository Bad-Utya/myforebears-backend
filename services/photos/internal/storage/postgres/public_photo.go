package postgres

import (
	"context"
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/domain/models"
	"github.com/google/uuid"
)

const photoColumns = `id, owner_user_id, tree_id, person_id, event_id, public_person_id, public_event_id,
	is_user_avatar, is_tree_avatar, is_person_avatar, file_name, mime_type, size_bytes, object_key, created_at`

func (s *Storage) ListPublicPersonPhotos(ctx context.Context, publicPersonID uuid.UUID) ([]models.Photo, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+photoColumns+` FROM photos WHERE public_person_id=$1 ORDER BY created_at`, publicPersonID)
	if err != nil {
		return nil, fmt.Errorf("storage.postgres.ListPublicPersonPhotos: %w", err)
	}
	defer rows.Close()
	result := make([]models.Photo, 0)
	for rows.Next() {
		photo, err := scanPhotoFromRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, photo)
	}
	return result, rows.Err()
}

func (s *Storage) UnsetPublicPersonAvatar(ctx context.Context, publicPersonID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `UPDATE photos SET is_person_avatar=FALSE WHERE public_person_id=$1 AND is_person_avatar=TRUE`, publicPersonID)
	if err != nil {
		return fmt.Errorf("storage.postgres.UnsetPublicPersonAvatar: %w", err)
	}
	return nil
}

func (s *Storage) DeletePublicPersonMedia(ctx context.Context, publicPersonID uuid.UUID) ([]models.Photo, error) {
	rows, err := s.pool.Query(ctx, `DELETE FROM photos WHERE public_person_id=$1 RETURNING `+photoColumns, publicPersonID)
	if err != nil {
		return nil, fmt.Errorf("storage.postgres.DeletePublicPersonMedia: %w", err)
	}
	defer rows.Close()
	result := make([]models.Photo, 0)
	for rows.Next() {
		photo, err := scanPhotoFromRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, photo)
	}
	return result, rows.Err()
}
