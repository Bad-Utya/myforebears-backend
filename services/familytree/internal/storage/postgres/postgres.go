package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
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

func (s *Storage) CreatePerson(ctx context.Context, person models.Person) error {
	const op = "storage.postgres.CreatePerson"

	_, err := s.pool.Exec(
		ctx,
		`INSERT INTO persons (id, tree_id, first_name, last_name, patronymic, gender, avatar_photo_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		person.ID,
		person.TreeID,
		person.FirstName,
		person.LastName,
		person.Patronymic,
		person.Gender,
		person.AvatarPhotoID,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) CreateTree(ctx context.Context, tree models.Tree) error {
	const op = "storage.postgres.CreateTree"

	_, err := s.pool.Exec(
		ctx,
		`INSERT INTO trees (id, creator_id, is_view_restricted, is_public_on_main_page, name)
		 VALUES ($1, $2, $3, $4, $5)`,
		tree.ID,
		tree.CreatorID,
		tree.IsViewRestricted,
		tree.IsPublicOnMainPage,
		tree.Name,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetTree(ctx context.Context, treeID uuid.UUID) (models.Tree, error) {
	const op = "storage.postgres.GetTree"

	var tree models.Tree
	err := s.pool.QueryRow(
		ctx,
		`SELECT id, creator_id, created_at, is_view_restricted, is_public_on_main_page, name
		 FROM trees WHERE id = $1`,
		treeID,
	).Scan(&tree.ID, &tree.CreatorID, &tree.CreatedAt, &tree.IsViewRestricted, &tree.IsPublicOnMainPage, &tree.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Tree{}, fmt.Errorf("%s: %w", op, storage.ErrTreeNotFound)
		}
		return models.Tree{}, fmt.Errorf("%s: %w", op, err)
	}

	return tree, nil
}

func (s *Storage) UpdateTreeSettings(ctx context.Context, treeID uuid.UUID, isViewRestricted bool, isPublicOnMainPage bool, name string) error {
	const op = "storage.postgres.UpdateTreeSettings"

	cmdTag, err := s.pool.Exec(
		ctx,
		`UPDATE trees
		 SET is_view_restricted = $1,
		     is_public_on_main_page = $2,
		     name = $3
		 WHERE id = $4`,
		isViewRestricted,
		isPublicOnMainPage,
		name,
		treeID,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrTreeNotFound)
	}

	return nil
}

func (s *Storage) AddTreeAccessEmail(ctx context.Context, treeID uuid.UUID, email string) error {
	const op = "storage.postgres.AddTreeAccessEmail"

	_, err := s.pool.Exec(
		ctx,
		`INSERT INTO tree_access_emails (tree_id, email)
		 VALUES ($1, $2)`,
		treeID,
		email,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%s: %w", op, storage.ErrTreeAccessEmailExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ListTreeAccessEmails(ctx context.Context, treeID uuid.UUID) ([]string, error) {
	const op = "storage.postgres.ListTreeAccessEmails"

	rows, err := s.pool.Query(
		ctx,
		`SELECT email
		 FROM tree_access_emails
		 WHERE tree_id = $1
		 ORDER BY email ASC`,
		treeID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	emails := make([]string, 0)
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		emails = append(emails, email)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return emails, nil
}

func (s *Storage) IsTreeAccessEmailAllowed(ctx context.Context, treeID uuid.UUID, email string) (bool, error) {
	const op = "storage.postgres.IsTreeAccessEmailAllowed"

	var allowed bool
	err := s.pool.QueryRow(
		ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM tree_access_emails
			WHERE tree_id = $1 AND email = $2
		)`,
		treeID,
		email,
	).Scan(&allowed)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return allowed, nil
}

func (s *Storage) DeleteTreeAccessEmail(ctx context.Context, treeID uuid.UUID, email string) error {
	const op = "storage.postgres.DeleteTreeAccessEmail"

	cmdTag, err := s.pool.Exec(
		ctx,
		`DELETE FROM tree_access_emails
		 WHERE tree_id = $1 AND email = $2`,
		treeID,
		email,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrTreeAccessEmailNotFound)
	}

	return nil
}

func (s *Storage) GetTreesByCreator(ctx context.Context, creatorID int) ([]models.Tree, error) {
	const op = "storage.postgres.GetTreesByCreator"

	rows, err := s.pool.Query(
		ctx,
		`SELECT id, creator_id, created_at, is_view_restricted, is_public_on_main_page, name
		 FROM trees
		 WHERE creator_id = $1
		 ORDER BY created_at DESC`,
		creatorID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	trees := make([]models.Tree, 0)
	for rows.Next() {
		var tree models.Tree
		if err := rows.Scan(&tree.ID, &tree.CreatorID, &tree.CreatedAt, &tree.IsViewRestricted, &tree.IsPublicOnMainPage, &tree.Name); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		trees = append(trees, tree)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return trees, nil
}

func (s *Storage) GetPublicTreesByCreator(ctx context.Context, creatorID int) ([]models.Tree, error) {
	const op = "storage.postgres.GetPublicTreesByCreator"

	rows, err := s.pool.Query(
		ctx,
		`SELECT id, creator_id, created_at, is_view_restricted, is_public_on_main_page, name
		 FROM trees
		 WHERE creator_id = $1 AND is_public_on_main_page = TRUE
		 ORDER BY created_at DESC`,
		creatorID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	trees := make([]models.Tree, 0)
	for rows.Next() {
		var tree models.Tree
		if err := rows.Scan(&tree.ID, &tree.CreatorID, &tree.CreatedAt, &tree.IsViewRestricted, &tree.IsPublicOnMainPage, &tree.Name); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		trees = append(trees, tree)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return trees, nil
}

func (s *Storage) GetRandomPublicTrees(ctx context.Context, limit int) ([]models.Tree, error) {
	const op = "storage.postgres.GetRandomPublicTrees"

	rows, err := s.pool.Query(
		ctx,
		`SELECT id, creator_id, created_at, is_view_restricted, is_public_on_main_page, name
		 FROM trees
		 WHERE is_public_on_main_page = TRUE
		 ORDER BY RANDOM()
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	trees := make([]models.Tree, 0)
	for rows.Next() {
		var tree models.Tree
		if err := rows.Scan(&tree.ID, &tree.CreatorID, &tree.CreatedAt, &tree.IsViewRestricted, &tree.IsPublicOnMainPage, &tree.Name); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		trees = append(trees, tree)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return trees, nil
}

func (s *Storage) GetPerson(ctx context.Context, personID uuid.UUID) (models.Person, error) {
	const op = "storage.postgres.GetPerson"

	var person models.Person
	err := s.pool.QueryRow(
		ctx,
		`SELECT id, tree_id, first_name, last_name, COALESCE(patronymic, ''), gender, avatar_photo_id
		 FROM persons WHERE id = $1`,
		personID,
	).Scan(
		&person.ID,
		&person.TreeID,
		&person.FirstName,
		&person.LastName,
		&person.Patronymic,
		&person.Gender,
		&person.AvatarPhotoID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Person{}, fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
		}
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	return person, nil
}

func (s *Storage) UpdatePerson(ctx context.Context, person models.Person) error {
	const op = "storage.postgres.UpdatePerson"

	cmdTag, err := s.pool.Exec(
		ctx,
		`UPDATE persons
		 SET first_name = $1,
		     last_name = $2,
		     patronymic = $3,
		     gender = $4,
		     updated_at = NOW()
		 WHERE id = $5`,
		person.FirstName,
		person.LastName,
		person.Patronymic,
		person.Gender,
		person.ID,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
	}

	return nil
}

func (s *Storage) UpdatePersonAvatarPhoto(ctx context.Context, personID uuid.UUID, avatarPhotoID *uuid.UUID) error {
	const op = "storage.postgres.UpdatePersonAvatarPhoto"

	cmdTag, err := s.pool.Exec(
		ctx,
		`UPDATE persons
		 SET avatar_photo_id = $1,
		     updated_at = NOW()
		 WHERE id = $2`,
		avatarPhotoID,
		personID,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
	}

	return nil
}

func (s *Storage) DeletePerson(ctx context.Context, personID uuid.UUID) error {
	const op = "storage.postgres.DeletePerson"

	cmdTag, err := s.pool.Exec(ctx, `DELETE FROM persons WHERE id = $1`, personID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
	}

	return nil
}

func (s *Storage) GetPersonsByTree(ctx context.Context, treeID uuid.UUID) ([]models.Person, error) {
	const op = "storage.postgres.GetPersonsByTree"

	rows, err := s.pool.Query(
		ctx,
		`SELECT id, tree_id, first_name, last_name, COALESCE(patronymic, ''), gender, avatar_photo_id
		 FROM persons WHERE tree_id = $1`,
		treeID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	persons := make([]models.Person, 0)
	for rows.Next() {
		var person models.Person
		if err := rows.Scan(
			&person.ID,
			&person.TreeID,
			&person.FirstName,
			&person.LastName,
			&person.Patronymic,
			&person.Gender,
			&person.AvatarPhotoID,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		persons = append(persons, person)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return persons, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}
