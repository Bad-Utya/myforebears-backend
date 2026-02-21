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

// SaveLink saves a password reset link hash associated with the email in Redis with the given TTL.
func (s *Storage) SaveLink(ctx context.Context, email string, linkHash string, ttl time.Duration) error {
	const op = "storage.redis.SaveLink"

	key := "reset_link:" + linkHash
	if err := s.client.Set(ctx, key, email, ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetEmailByLink retrieves the email associated with the given password reset link hash from Redis.
func (s *Storage) GetEmailByLink(ctx context.Context, linkHash string) (string, error) {
	const op = "storage.redis.GetEmailByLink"

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

// verifyCodeScript atomically verifies the input code hash against the stored
// one and acts accordingly — all in a single Lua transaction:
//   - KEYS[1]: Redis hash key for the email
//   - ARGV[1]: SHA-256 hex of the code entered by the user
//
// Returns {pass_hash, "1"} on a match (key is DELeted inside the script).
// Returns {pass_hash, "0"} on a mismatch (key DELeted when exhausted).
// Returns an error reply for missing key or exhausted attempts.
var verifyCodeScript = redis.NewScript(`
local key        = KEYS[1]
local input_hash = ARGV[1]

local pass_hash = redis.call('HGET', key, 'pass_hash')
if pass_hash == false then
  return redis.error_reply('CODE_NOT_FOUND')
end

local attempts = tonumber(redis.call('HGET', key, 'attempts'))
if attempts <= 0 then
  redis.call('DEL', key)
  return redis.error_reply('NO_ATTEMPTS_LEFT')
end

local stored_hash = redis.call('HGET', key, 'code_hash')
if stored_hash == input_hash then
  redis.call('DEL', key)
  return {pass_hash, "1"}
end

redis.call('HINCRBY', key, 'attempts', -1)
return {pass_hash, "0"}
`)

// VerifyCode atomically checks inputCodeHash against the stored hash in Redis.
// On a match the key is deleted inside the Lua script.
// On a mismatch attempts are decremented (key deleted when exhausted).
func (s *Storage) VerifyCode(
	ctx context.Context,
	email string,
	inputCodeHash string,
) (passHash []byte, matched bool, err error) {
	const op = "storage.redis.VerifyCode"

	key := verifyKey(email)

	res, err := verifyCodeScript.Run(ctx, s.client, []string{key}, inputCodeHash).Slice()
	if err != nil {
		if err.Error() == "CODE_NOT_FOUND" {
			return nil, false, storage.ErrCodeNotFound
		}
		if err.Error() == "NO_ATTEMPTS_LEFT" {
			return nil, false, storage.ErrNoAttemptsLeft
		}
		return nil, false, fmt.Errorf("%s: %w", op, err)
	}

	passHash, err = hex.DecodeString(res[0].(string))
	if err != nil {
		return nil, false, fmt.Errorf("%s: failed to decode pass_hash: %w", op, err)
	}

	matched = res[1].(string) == "1"

	return passHash, matched, nil
}

// DeleteCode removes the verification entry from Redis.
func (s *Storage) DeleteCode(ctx context.Context, email string) error {
	const op = "storage.redis.DeleteCode"

	if err := s.client.Del(ctx, verifyKey(email)).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// BlacklistToken adds the given token to Redis with a value of "1" and the specified TTL.
func (s *Storage) BlacklistToken(ctx context.Context, token string, ttl time.Duration) error {
	const op = "storage.redis.BlacklistToken"

	key := "blacklist:" + token
	if err := s.client.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// BlacklistEmail saves the current Unix timestamp under "blacklist:<email>" with the given TTL.
// Any token whose created_at is less than this timestamp is considered revoked.
func (s *Storage) BlacklistEmail(ctx context.Context, email string, ttl time.Duration) error {
	const op = "storage.redis.BlacklistEmail"

	key := "blacklist:" + email
	logoutAt := time.Now().Unix()
	if err := s.client.Set(ctx, key, strconv.FormatInt(logoutAt, 10), ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Close() {
	s.client.Close()
}
