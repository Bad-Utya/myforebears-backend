package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) CreatePublicPerson(ctx context.Context, person models.PublicPerson) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("storage.postgres.CreatePublicPerson: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var gender any
	if person.Gender != "" {
		gender = person.Gender
	}
	_, err = tx.Exec(ctx, `INSERT INTO public_persons
		(id, owner_user_id, first_name, last_name, patronymic, gender, biography, avatar_photo_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, person.ID, person.OwnerUserID, person.FirstName,
		person.LastName, person.Patronymic, gender, person.Biography, person.AvatarPhotoID, person.CreatedAt, person.UpdatedAt)
	if err != nil {
		return fmt.Errorf("storage.postgres.CreatePublicPerson: %w", err)
	}
	if err := insertPublicEvents(ctx, tx, person.ID, person.Events); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("storage.postgres.CreatePublicPerson: %w", err)
	}
	return nil
}

func (s *Storage) GetPublicPerson(ctx context.Context, personID uuid.UUID) (models.PublicPerson, error) {
	var p models.PublicPerson
	var gender *string
	err := s.pool.QueryRow(ctx, `SELECT id, owner_user_id, first_name, last_name, patronymic, gender::text,
		biography, avatar_photo_id, created_at, updated_at FROM public_persons WHERE id=$1`, personID).
		Scan(&p.ID, &p.OwnerUserID, &p.FirstName, &p.LastName, &p.Patronymic, &gender, &p.Biography, &p.AvatarPhotoID, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, storage.ErrPublicPersonNotFound
	}
	if err != nil {
		return p, fmt.Errorf("storage.postgres.GetPublicPerson: %w", err)
	}
	if gender != nil {
		p.Gender = models.Gender(*gender)
	}
	events, err := s.listPublicEvents(ctx, p.ID)
	if err != nil {
		return models.PublicPerson{}, err
	}
	p.Events = events
	p.Tags, err = s.listPublicPersonTags(ctx, p.ID)
	if err != nil {
		return models.PublicPerson{}, err
	}
	return p, nil
}

func (s *Storage) ListRandomPublicPersons(ctx context.Context, limit int) ([]models.PublicPerson, error) {
	return s.listPublicPersons(ctx, `SELECT id FROM public_persons ORDER BY random() LIMIT $1`, limit)
}

func (s *Storage) ListPublicPersonsByOwner(ctx context.Context, ownerUserID int, limit int) ([]models.PublicPerson, error) {
	return s.listPublicPersons(ctx, `SELECT id FROM public_persons WHERE owner_user_id=$1 ORDER BY created_at DESC LIMIT $2`, ownerUserID, limit)
}

func (s *Storage) SearchPublicPersons(ctx context.Context, query string, limit int) ([]models.PublicPerson, error) {
	pattern := "%" + query + "%"
	return s.listPublicPersons(ctx, `SELECT id FROM public_persons
		WHERE concat_ws(' ', first_name, last_name, patronymic, biography) ILIKE $1
		ORDER BY updated_at DESC LIMIT $2`, pattern, limit)
}

func (s *Storage) listPublicPersons(ctx context.Context, query string, args ...any) ([]models.PublicPerson, error) {
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("storage.postgres.listPublicPersons: %w", err)
	}
	defer rows.Close()
	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	result := make([]models.PublicPerson, 0, len(ids))
	for _, id := range ids {
		p, err := s.GetPublicPerson(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func (s *Storage) UpdatePublicPerson(ctx context.Context, person models.PublicPerson) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var gender any
	if person.Gender != "" {
		gender = person.Gender
	}
	tag, err := tx.Exec(ctx, `UPDATE public_persons SET first_name=$1,last_name=$2,patronymic=$3,
		gender=$4,biography=$5,updated_at=$6 WHERE id=$7`, person.FirstName, person.LastName, person.Patronymic,
		gender, person.Biography, person.UpdatedAt, person.ID)
	if err != nil {
		return fmt.Errorf("storage.postgres.UpdatePublicPerson: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return storage.ErrPublicPersonNotFound
	}
	if _, err := tx.Exec(ctx, `DELETE FROM public_person_events WHERE public_person_id=$1`, person.ID); err != nil {
		return err
	}
	if err := insertPublicEvents(ctx, tx, person.ID, person.Events); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Storage) SetPublicPersonAvatarPhoto(ctx context.Context, personID uuid.UUID, avatarPhotoID *uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `UPDATE public_persons SET avatar_photo_id=$1,updated_at=NOW() WHERE id=$2`, avatarPhotoID, personID)
	if err != nil {
		return fmt.Errorf("storage.postgres.SetPublicPersonAvatarPhoto: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return storage.ErrPublicPersonNotFound
	}
	return nil
}

func (s *Storage) DeletePublicPerson(ctx context.Context, personID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM public_persons WHERE id=$1`, personID)
	if err != nil {
		return fmt.Errorf("storage.postgres.DeletePublicPerson: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return storage.ErrPublicPersonNotFound
	}
	return nil
}

func insertPublicEvents(ctx context.Context, tx pgx.Tx, personID uuid.UUID, events []models.PublicPersonEvent) error {
	for _, e := range events {
		_, err := tx.Exec(ctx, `INSERT INTO public_person_events
			(id,public_person_id,source_event_id,event_type_id,event_type_name,date_iso,date_precision,date_bound,date_unknown)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, e.ID, personID, e.SourceEventID, e.EventTypeID, e.EventTypeName,
			e.DateISO, e.DatePrecision, e.DateBound, e.DateUnknown)
		if err != nil {
			return fmt.Errorf("storage.postgres.insertPublicEvents: %w", err)
		}
	}
	return nil
}

func (s *Storage) listPublicEvents(ctx context.Context, personID uuid.UUID) ([]models.PublicPersonEvent, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,public_person_id,source_event_id,event_type_id,event_type_name,
		date_iso,date_precision,date_bound,date_unknown FROM public_person_events WHERE public_person_id=$1 ORDER BY created_at,id`, personID)
	if err != nil {
		return nil, fmt.Errorf("storage.postgres.listPublicEvents: %w", err)
	}
	defer rows.Close()
	result := make([]models.PublicPersonEvent, 0)
	for rows.Next() {
		var e models.PublicPersonEvent
		if err := rows.Scan(&e.ID, &e.PublicPersonID, &e.SourceEventID, &e.EventTypeID, &e.EventTypeName, &e.DateISO, &e.DatePrecision, &e.DateBound, &e.DateUnknown); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}
