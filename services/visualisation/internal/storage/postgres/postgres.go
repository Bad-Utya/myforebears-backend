package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/storage"
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

func (s *Storage) CreateVisualisation(ctx context.Context, vis models.Visualisation) error {
	const op = "storage.postgres.CreateVisualisation"

	_, err := s.pool.Exec(
		ctx,
		`INSERT INTO visualisations (
			id, owner_user_id, tree_id, root_person_id, included_person_ids,
			type, status, file_name, mime_type, size_bytes, object_key,
			error_message, created_at, updated_at, completed_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15
		)`,
		vis.ID,
		vis.OwnerUserID,
		vis.TreeID,
		vis.RootPersonID,
		toStringIDs(vis.IncludedPersonIDs),
		string(vis.Type),
		string(vis.Status),
		vis.FileName,
		vis.MIMEType,
		vis.SizeBytes,
		vis.ObjectKey,
		vis.ErrorMessage,
		vis.CreatedAt,
		vis.UpdatedAt,
		vis.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetVisualisationByID(ctx context.Context, visualisationID uuid.UUID) (models.Visualisation, error) {
	const op = "storage.postgres.GetVisualisationByID"

	vis, err := s.scanVisualisation(
		ctx,
		`SELECT id, owner_user_id, tree_id, root_person_id, included_person_ids,
			type, status, file_name, mime_type, size_bytes, object_key,
			error_message, created_at, updated_at, completed_at
		 FROM visualisations
		 WHERE id = $1`,
		visualisationID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Visualisation{}, fmt.Errorf("%s: %w", op, storage.ErrVisualisationNotFound)
		}
		return models.Visualisation{}, fmt.Errorf("%s: %w", op, err)
	}

	return vis, nil
}

func (s *Storage) ListTreeVisualisations(ctx context.Context, treeID uuid.UUID) ([]models.Visualisation, error) {
	const op = "storage.postgres.ListTreeVisualisations"

	rows, err := s.pool.Query(
		ctx,
		`SELECT id, owner_user_id, tree_id, root_person_id, included_person_ids,
			type, status, file_name, mime_type, size_bytes, object_key,
			error_message, created_at, updated_at, completed_at
		 FROM visualisations
		 WHERE tree_id = $1
		 ORDER BY created_at DESC`,
		treeID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	items := make([]models.Visualisation, 0)
	for rows.Next() {
		vis, err := scanVisualisationFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		items = append(items, vis)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return items, nil
}

func (s *Storage) DeleteVisualisationByID(ctx context.Context, visualisationID uuid.UUID) (models.Visualisation, error) {
	const op = "storage.postgres.DeleteVisualisationByID"

	vis, err := s.scanVisualisation(
		ctx,
		`DELETE FROM visualisations
		 WHERE id = $1
		 RETURNING id, owner_user_id, tree_id, root_person_id, included_person_ids,
			type, status, file_name, mime_type, size_bytes, object_key,
			error_message, created_at, updated_at, completed_at`,
		visualisationID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Visualisation{}, fmt.Errorf("%s: %w", op, storage.ErrVisualisationNotFound)
		}
		return models.Visualisation{}, fmt.Errorf("%s: %w", op, err)
	}

	return vis, nil
}

func (s *Storage) SetVisualisationProcessing(ctx context.Context, visualisationID uuid.UUID) error {
	const op = "storage.postgres.SetVisualisationProcessing"

	_, err := s.pool.Exec(
		ctx,
		`UPDATE visualisations
		 SET status = $2,
		     error_message = '',
		     updated_at = NOW(),
		     completed_at = NULL
		 WHERE id = $1`,
		visualisationID,
		string(models.VisualisationStatusProcessing),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) SetVisualisationFailed(ctx context.Context, visualisationID uuid.UUID, errorMessage string) error {
	const op = "storage.postgres.SetVisualisationFailed"

	_, err := s.pool.Exec(
		ctx,
		`UPDATE visualisations
		 SET status = $2,
		     error_message = $3,
		     updated_at = NOW(),
		     completed_at = NOW()
		 WHERE id = $1`,
		visualisationID,
		string(models.VisualisationStatusFailed),
		errorMessage,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) SetVisualisationReady(ctx context.Context, visualisationID uuid.UUID, sizeBytes int64) error {
	const op = "storage.postgres.SetVisualisationReady"

	_, err := s.pool.Exec(
		ctx,
		`UPDATE visualisations
		 SET status = $2,
		     error_message = '',
		     size_bytes = $3,
		     updated_at = NOW(),
		     completed_at = NOW()
		 WHERE id = $1`,
		visualisationID,
		string(models.VisualisationStatusReady),
		sizeBytes,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) scanVisualisation(ctx context.Context, sql string, args ...any) (models.Visualisation, error) {
	row := s.pool.QueryRow(ctx, sql, args...)
	return scanVisualisationFromRow(row)
}

func scanVisualisationFromRow(row interface{ Scan(dest ...any) error }) (models.Visualisation, error) {
	var (
		vis               models.Visualisation
		visType           string
		visStatus         string
		includedPersonIDs []string
		completedAt       *time.Time
	)

	if err := row.Scan(
		&vis.ID,
		&vis.OwnerUserID,
		&vis.TreeID,
		&vis.RootPersonID,
		&includedPersonIDs,
		&visType,
		&visStatus,
		&vis.FileName,
		&vis.MIMEType,
		&vis.SizeBytes,
		&vis.ObjectKey,
		&vis.ErrorMessage,
		&vis.CreatedAt,
		&vis.UpdatedAt,
		&completedAt,
	); err != nil {
		return models.Visualisation{}, err
	}

	ids, err := parseUUIDStrings(includedPersonIDs)
	if err != nil {
		return models.Visualisation{}, err
	}

	vis.IncludedPersonIDs = ids
	vis.Type = models.VisualisationType(visType)
	vis.Status = models.VisualisationStatus(visStatus)
	vis.CompletedAt = completedAt

	return vis, nil
}

func parseUUIDStrings(values []string) ([]uuid.UUID, error) {
	if len(values) == 0 {
		return []uuid.UUID{}, nil
	}

	out := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		parsed, err := uuid.Parse(value)
		if err != nil {
			return nil, err
		}
		out = append(out, parsed)
	}

	return out, nil
}

func toStringIDs(values []uuid.UUID) []string {
	if len(values) == 0 {
		return []string{}
	}

	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value.String())
	}

	return out
}

func (s *Storage) Close() {
	s.pool.Close()
}
