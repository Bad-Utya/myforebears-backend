package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/events/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/events/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func (s *Storage) CreateEventType(ctx context.Context, eventType models.EventType) error {
	const op = "storage.postgres.CreateEventType"

	_, err := s.pool.Exec(
		ctx,
		`INSERT INTO event_types (id, owner_user_id, is_system, name, primary_persons_mode, primary_persons_count, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		eventType.ID,
		eventType.OwnerUserID,
		eventType.IsSystem,
		eventType.Name,
		eventType.PrimaryPersonsMode,
		eventType.PrimaryPersonsCount,
		eventType.CreatedAt,
		eventType.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%s: %w", op, storage.ErrEventTypeAlreadyExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetEventType(ctx context.Context, eventTypeID uuid.UUID) (models.EventType, error) {
	const op = "storage.postgres.GetEventType"

	var eventType models.EventType
	err := s.pool.QueryRow(
		ctx,
		`SELECT id, owner_user_id, is_system, name, primary_persons_mode, primary_persons_count, created_at, updated_at
		 FROM event_types
		 WHERE id = $1`,
		eventTypeID,
	).Scan(
		&eventType.ID,
		&eventType.OwnerUserID,
		&eventType.IsSystem,
		&eventType.Name,
		&eventType.PrimaryPersonsMode,
		&eventType.PrimaryPersonsCount,
		&eventType.CreatedAt,
		&eventType.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.EventType{}, fmt.Errorf("%s: %w", op, storage.ErrEventTypeNotFound)
		}
		return models.EventType{}, fmt.Errorf("%s: %w", op, err)
	}

	return eventType, nil
}

func (s *Storage) ListEventTypesForUser(ctx context.Context, requestUserID int) ([]models.EventType, error) {
	const op = "storage.postgres.ListEventTypesForUser"

	rows, err := s.pool.Query(
		ctx,
		`SELECT id, owner_user_id, is_system, name, primary_persons_mode, primary_persons_count, created_at, updated_at
		 FROM event_types
		 WHERE is_system = TRUE OR owner_user_id = $1
		 ORDER BY is_system DESC, lower(name) ASC`,
		requestUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	eventTypes := make([]models.EventType, 0)
	for rows.Next() {
		var eventType models.EventType
		if err := rows.Scan(
			&eventType.ID,
			&eventType.OwnerUserID,
			&eventType.IsSystem,
			&eventType.Name,
			&eventType.PrimaryPersonsMode,
			&eventType.PrimaryPersonsCount,
			&eventType.CreatedAt,
			&eventType.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		eventTypes = append(eventTypes, eventType)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return eventTypes, nil
}

func (s *Storage) DeleteEventType(ctx context.Context, eventTypeID uuid.UUID) error {
	const op = "storage.postgres.DeleteEventType"

	cmdTag, err := s.pool.Exec(ctx, `DELETE FROM event_types WHERE id = $1`, eventTypeID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrEventTypeNotFound)
	}

	return nil
}

func (s *Storage) HasEventsByType(ctx context.Context, eventTypeID uuid.UUID) (bool, error) {
	const op = "storage.postgres.HasEventsByType"

	var exists bool
	err := s.pool.QueryRow(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM events WHERE event_type_id = $1)`,
		eventTypeID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

func (s *Storage) CreateEvent(ctx context.Context, event models.Event) error {
	const op = "storage.postgres.CreateEvent"

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`INSERT INTO events (id, tree_id, event_type_id, date_value, date_precision, date_bound, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		event.ID,
		event.TreeID,
		event.EventTypeID,
		event.DateValue,
		event.DatePrecision,
		event.DateBound,
		event.CreatedAt,
		event.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := insertParticipants(ctx, tx, event.ID, event.PrimaryPersonIDs, true); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := insertParticipants(ctx, tx, event.ID, event.AdditionalPersonIDs, false); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetEvent(ctx context.Context, eventID uuid.UUID) (models.Event, error) {
	const op = "storage.postgres.GetEvent"

	var event models.Event
	err := s.pool.QueryRow(
		ctx,
		`SELECT id, tree_id, event_type_id, date_value, date_precision, date_bound, created_at, updated_at
		 FROM events
		 WHERE id = $1`,
		eventID,
	).Scan(
		&event.ID,
		&event.TreeID,
		&event.EventTypeID,
		&event.DateValue,
		&event.DatePrecision,
		&event.DateBound,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Event{}, fmt.Errorf("%s: %w", op, storage.ErrEventNotFound)
		}
		return models.Event{}, fmt.Errorf("%s: %w", op, err)
	}

	event.PrimaryPersonIDs, err = selectParticipants(ctx, s.pool, event.ID, true)
	if err != nil {
		return models.Event{}, fmt.Errorf("%s: %w", op, err)
	}

	event.AdditionalPersonIDs, err = selectParticipants(ctx, s.pool, event.ID, false)
	if err != nil {
		return models.Event{}, fmt.Errorf("%s: %w", op, err)
	}

	return event, nil
}

func (s *Storage) ListEventsByTree(ctx context.Context, treeID uuid.UUID) ([]models.Event, error) {
	const op = "storage.postgres.ListEventsByTree"

	rows, err := s.pool.Query(
		ctx,
		`SELECT id, tree_id, event_type_id, date_value, date_precision, date_bound, created_at, updated_at
		 FROM events
		 WHERE tree_id = $1
		 ORDER BY date_value ASC, created_at ASC`,
		treeID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	events := make([]models.Event, 0)
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(
			&event.ID,
			&event.TreeID,
			&event.EventTypeID,
			&event.DateValue,
			&event.DatePrecision,
			&event.DateBound,
			&event.CreatedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		event.PrimaryPersonIDs, err = selectParticipants(ctx, s.pool, event.ID, true)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		event.AdditionalPersonIDs, err = selectParticipants(ctx, s.pool, event.ID, false)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return events, nil
}

func (s *Storage) UpdateEvent(ctx context.Context, event models.Event) error {
	const op = "storage.postgres.UpdateEvent"

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	cmdTag, err := tx.Exec(
		ctx,
		`UPDATE events
		 SET event_type_id = $1,
		     date_value = $2,
		     date_precision = $3,
		     date_bound = $4,
		     updated_at = $5
		 WHERE id = $6`,
		event.EventTypeID,
		event.DateValue,
		event.DatePrecision,
		event.DateBound,
		event.UpdatedAt,
		event.ID,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrEventNotFound)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM event_primary_persons WHERE event_id = $1`, event.ID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM event_additional_persons WHERE event_id = $1`, event.ID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := insertParticipants(ctx, tx, event.ID, event.PrimaryPersonIDs, true); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := insertParticipants(ctx, tx, event.ID, event.AdditionalPersonIDs, false); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	const op = "storage.postgres.DeleteEvent"

	cmdTag, err := s.pool.Exec(ctx, `DELETE FROM events WHERE id = $1`, eventID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrEventNotFound)
	}

	return nil
}

func (s *Storage) Close() {
	s.pool.Close()
}

func insertParticipants(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, personIDs []uuid.UUID, primary bool) error {
	query := `INSERT INTO event_additional_persons (event_id, person_id, position) VALUES ($1, $2, $3)`
	if primary {
		query = `INSERT INTO event_primary_persons (event_id, person_id, position) VALUES ($1, $2, $3)`
	}

	for idx, personID := range personIDs {
		if _, err := tx.Exec(ctx, query, eventID, personID, idx+1); err != nil {
			return err
		}
	}

	return nil
}

type participantQueryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func selectParticipants(ctx context.Context, q participantQueryer, eventID uuid.UUID, primary bool) ([]uuid.UUID, error) {
	query := `SELECT person_id FROM event_additional_persons WHERE event_id = $1 ORDER BY position ASC`
	if primary {
		query = `SELECT person_id FROM event_primary_persons WHERE event_id = $1 ORDER BY position ASC`
	}

	rows, err := q.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var personID uuid.UUID
		if err := rows.Scan(&personID); err != nil {
			return nil, err
		}
		ids = append(ids, personID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}
