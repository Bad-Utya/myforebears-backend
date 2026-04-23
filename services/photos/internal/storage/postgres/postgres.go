package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(host string, port int, user string, password string, dbname string) (*Storage, error) {
	const op = "storage.postgres.New"

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user,
		password,
		host,
		port,
		dbname,
	)

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) CreatePhoto(ctx context.Context, photo models.Photo) error {
	const op = "storage.postgres.CreatePhoto"

	_, err := s.pool.Exec(
		ctx,
		`INSERT INTO photos (
			id, owner_user_id, tree_id, person_id, event_id, is_user_avatar, is_person_avatar,
			file_name, mime_type, size_bytes, object_key, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12
		)`,
		photo.ID,
		photo.OwnerUserID,
		photo.TreeID,
		photo.PersonID,
		photo.EventID,
		photo.IsUserAvatar,
		photo.IsPersonAvatar,
		photo.FileName,
		photo.MIMEType,
		photo.SizeBytes,
		photo.ObjectKey,
		photo.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetPhotoByID(ctx context.Context, photoID uuid.UUID) (models.Photo, error) {
	const op = "storage.postgres.GetPhotoByID"

	photo, err := s.scanPhoto(
		ctx,
		`SELECT id, owner_user_id, tree_id, person_id, event_id, is_user_avatar, is_person_avatar,
			file_name, mime_type, size_bytes, object_key, created_at
		 FROM photos
		 WHERE id = $1`,
		photoID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Photo{}, fmt.Errorf("%s: %w", op, storage.ErrPhotoNotFound)
		}
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	return photo, nil
}

func (s *Storage) GetUserAvatar(ctx context.Context, ownerUserID int) (models.Photo, error) {
	const op = "storage.postgres.GetUserAvatar"

	photo, err := s.scanPhoto(
		ctx,
		`SELECT id, owner_user_id, tree_id, person_id, event_id, is_user_avatar, is_person_avatar,
			file_name, mime_type, size_bytes, object_key, created_at
		 FROM photos
		 WHERE owner_user_id = $1 AND is_user_avatar = TRUE
		 LIMIT 1`,
		ownerUserID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Photo{}, fmt.Errorf("%s: %w", op, storage.ErrPhotoNotFound)
		}
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	return photo, nil
}

func (s *Storage) GetPersonAvatar(ctx context.Context, personID uuid.UUID) (models.Photo, error) {
	const op = "storage.postgres.GetPersonAvatar"

	photo, err := s.scanPhoto(
		ctx,
		`SELECT id, owner_user_id, tree_id, person_id, event_id, is_user_avatar, is_person_avatar,
			file_name, mime_type, size_bytes, object_key, created_at
		 FROM photos
		 WHERE person_id = $1 AND is_person_avatar = TRUE
		 LIMIT 1`,
		personID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Photo{}, fmt.Errorf("%s: %w", op, storage.ErrPhotoNotFound)
		}
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	return photo, nil
}

func (s *Storage) UnsetPersonAvatar(ctx context.Context, personID uuid.UUID) error {
	const op = "storage.postgres.UnsetPersonAvatar"

	_, err := s.pool.Exec(
		ctx,
		`UPDATE photos
		 SET is_person_avatar = FALSE
		 WHERE person_id = $1 AND is_person_avatar = TRUE`,
		personID,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ListPersonPhotos(ctx context.Context, personID uuid.UUID) ([]models.Photo, error) {
	const op = "storage.postgres.ListPersonPhotos"

	rows, err := s.pool.Query(
		ctx,
		`SELECT id, owner_user_id, tree_id, person_id, event_id, is_user_avatar, is_person_avatar,
			file_name, mime_type, size_bytes, object_key, created_at
		 FROM photos
		 WHERE person_id = $1
		 ORDER BY created_at DESC`,
		personID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	photos := make([]models.Photo, 0)
	for rows.Next() {
		photo, err := scanPhotoFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		photos = append(photos, photo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return photos, nil
}

func (s *Storage) ListEventPhotos(ctx context.Context, eventID uuid.UUID) ([]models.Photo, error) {
	const op = "storage.postgres.ListEventPhotos"

	rows, err := s.pool.Query(
		ctx,
		`SELECT id, owner_user_id, tree_id, person_id, event_id, is_user_avatar, is_person_avatar,
			file_name, mime_type, size_bytes, object_key, created_at
		 FROM photos
		 WHERE event_id = $1
		 ORDER BY created_at DESC`,
		eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	photos := make([]models.Photo, 0)
	for rows.Next() {
		photo, err := scanPhotoFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		photos = append(photos, photo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return photos, nil
}

func (s *Storage) DeletePhotoByID(ctx context.Context, photoID uuid.UUID) (models.Photo, error) {
	const op = "storage.postgres.DeletePhotoByID"

	photo, err := s.scanPhoto(
		ctx,
		`DELETE FROM photos
		 WHERE id = $1
		 RETURNING id, owner_user_id, tree_id, person_id, event_id, is_user_avatar, is_person_avatar,
			file_name, mime_type, size_bytes, object_key, created_at`,
		photoID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Photo{}, fmt.Errorf("%s: %w", op, storage.ErrPhotoNotFound)
		}
		return models.Photo{}, fmt.Errorf("%s: %w", op, err)
	}

	return photo, nil
}

func (s *Storage) scanPhoto(ctx context.Context, sql string, args ...any) (models.Photo, error) {
	row := s.pool.QueryRow(ctx, sql, args...)
	return scanPhotoFromRow(row)
}

func scanPhotoFromRow(row interface {
	Scan(dest ...any) error
}) (models.Photo, error) {
	var (
		photo    models.Photo
		treeID   *uuid.UUID
		personID *uuid.UUID
		eventID  *uuid.UUID
	)

	if err := row.Scan(
		&photo.ID,
		&photo.OwnerUserID,
		&treeID,
		&personID,
		&eventID,
		&photo.IsUserAvatar,
		&photo.IsPersonAvatar,
		&photo.FileName,
		&photo.MIMEType,
		&photo.SizeBytes,
		&photo.ObjectKey,
		&photo.CreatedAt,
	); err != nil {
		return models.Photo{}, err
	}

	photo.TreeID = treeID
	photo.PersonID = personID
	photo.EventID = eventID

	return photo, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}
