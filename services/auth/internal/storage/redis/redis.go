package redis

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/storage"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	client *redis.Client
}

func New(addr string, password string, db int) (*Storage, error) {
	const op = "storage.redis.New"

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{client: client}, nil
}

func verifyKey(email string) string {
	// Namespace verification entries by email.
	return "verify:" + email
}

// SaveCode saves email, password hash, code hash, attempts count
// and creation timestamp to Redis with the given TTL.
func (s *Storage) SaveCode(
	ctx context.Context,
	email string,
	passHash []byte,
	codeHash string,
	attempts int,
	ttl time.Duration,
) error {
	const op = "storage.redis.SaveCode"

	key := verifyKey(email)

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"pass_hash":  hex.EncodeToString(passHash),
		"code_hash":  codeHash,
		"attempts":   attempts,
		"created_at": time.Now().Unix(),
	})
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetCode retrieves verification data from Redis for the given email.
func (s *Storage) GetCode(
	ctx context.Context,
	email string,
) (passHash []byte, codeHash string, attempts int, createdAt time.Time, err error) {
	const op = "storage.redis.GetCode"

	key := verifyKey(email)

	data, err := s.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, "", 0, time.Time{}, fmt.Errorf("%s: %w", op, err)
	}

	if len(data) == 0 {
		return nil, "", 0, time.Time{}, storage.ErrCodeNotFound
	}

	passHash, err = hex.DecodeString(data["pass_hash"])
	if err != nil {
		return nil, "", 0, time.Time{}, fmt.Errorf("%s: failed to decode pass_hash: %w", op, err)
	}

	codeHash = data["code_hash"]

	attempts, err = strconv.Atoi(data["attempts"])
	if err != nil {
		return nil, "", 0, time.Time{}, fmt.Errorf("%s: failed to parse attempts: %w", op, err)
	}

	createdAtUnix, err := strconv.ParseInt(data["created_at"], 10, 64)
	if err != nil {
		return nil, "", 0, time.Time{}, fmt.Errorf("%s: failed to parse created_at: %w", op, err)
	}
	createdAt = time.Unix(createdAtUnix, 0)

	return passHash, codeHash, attempts, createdAt, nil
}

func (s *Storage) SaveLink(ctx context.Context, email string, linkHash string, ttl time.Duration) error {
	const op = "storage.redis.SaveLink"

	key := "reset_link:" + linkHash
	if err := s.client.Set(ctx, key, email, ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetLink(ctx context.Context, linkHash string) (string, error) {
	const op = "storage.redis.GetLink"

	key := "reset_link:" + linkHash
	email, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", storage.ErrLinkNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return email, nil
}

// DecrementAttempts decreases the remaining attempts by 1 and returns the new value.
func (s *Storage) DecrementAttempts(ctx context.Context, email string) (int, error) {
	const op = "storage.redis.DecrementAttempts"

	key := verifyKey(email)

	val, err := s.client.HIncrBy(ctx, key, "attempts", -1).Result()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return int(val), nil
}

// DeleteCode removes the verification entry from Redis.
func (s *Storage) DeleteCode(ctx context.Context, email string) error {
	const op = "storage.redis.DeleteCode"

	if err := s.client.Del(ctx, verifyKey(email)).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Close() {
	s.client.Close()
}
