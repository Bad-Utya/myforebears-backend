package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/storage"
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
		user, password, host, port, dbname,
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

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte, nickname string) (int, error) {
	const op = "storage.postgres.SaveUser"

	var id int
	err := s.pool.QueryRow(ctx,
		"INSERT INTO users (email, pass_hash, nickname) VALUES ($1, $2, $3) RETURNING id",
		email, passHash, nickname,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "duplicate key value violates unique constraint" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetUser(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.GetUser"

	var user models.User
	err := s.pool.QueryRow(ctx,
		"SELECT id, email, pass_hash, nickname FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PassHash, &user.Nickname)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) UpdatePassword(ctx context.Context, email string, newPassHash []byte) error {
	const op = "storage.postgres.UpdatePassword"

	cmdTag, err := s.pool.Exec(ctx,
		"UPDATE users SET pass_hash = $1 WHERE email = $2",
		newPassHash, email,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
	}

	return nil
}

func (s *Storage) GetUserByID(ctx context.Context, userID int) (models.User, error) {
	const op = "storage.postgres.GetUserByID"

	var user models.User
	err := s.pool.QueryRow(ctx,
		"SELECT id, email, pass_hash, nickname FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.PassHash, &user.Nickname)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) UpdateNickname(ctx context.Context, userID int, nickname string) error {
	const op = "storage.postgres.UpdateNickname"

	cmdTag, err := s.pool.Exec(ctx,
		"UPDATE users SET nickname = $1 WHERE id = $2",
		nickname, userID,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
	}

	return nil
}

func (s *Storage) Close() {
	s.pool.Close()
}
