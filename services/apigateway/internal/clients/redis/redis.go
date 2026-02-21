package redis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func New(addr string, password string, db int) (*Client, error) {
	const op = "storage.redis.New"

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{client: client}, nil
}

// IsTokenBlacklisted checks whether the given token is revoked.
// It checks two conditions:
//  1. blacklist:<token>  — set on Logout (individual token revocation).
//  2. blacklist:<email>  — set on LogoutFromAllDevices; the stored value is the
//     Unix timestamp of the logout event. A token is revoked when its
//     created_at is strictly less than that timestamp.
func (c *Client) IsTokenBlacklisted(ctx context.Context, token string, email string, createdAt int64) (bool, error) {
	const op = "clients.redis.IsTokenBlacklisted"

	// 1. Individual token revocation.
	exists, err := c.client.Exists(ctx, "blacklist:"+token).Result()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if exists > 0 {
		return true, nil
	}

	// 2. Global email revocation (LogoutFromAllDevices).
	timeStr, err := c.client.Get(ctx, "blacklist:"+email).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	logoutAt, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("%s: failed to parse logout timestamp: %w", op, err)
	}

	// Token was issued before the global logout → revoked.
	if createdAt < logoutAt {
		return true, nil
	}

	return false, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
