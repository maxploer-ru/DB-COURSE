package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const refreshSessionKeyPrefix = "auth:refresh:"

type RefreshSessionCache struct {
	client *redis.Client
}

func NewRefreshSessionCache(client *redis.Client) *RefreshSessionCache {
	return &RefreshSessionCache{client: client}
}

func (c *RefreshSessionCache) Save(ctx context.Context, tokenID string, userID int, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return fmt.Errorf("refresh session ttl is already expired")
	}
	return c.client.Set(ctx, refreshSessionKeyPrefix+tokenID, strconv.Itoa(userID), ttl).Err()
}

func (c *RefreshSessionCache) GetUserID(ctx context.Context, tokenID string) (int, bool, error) {
	val, err := c.client.Get(ctx, refreshSessionKeyPrefix+tokenID).Result()
	if errors.Is(err, redis.Nil) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	userID, convErr := strconv.Atoi(val)
	if convErr != nil {
		return 0, false, fmt.Errorf("parse refresh session user id: %w", convErr)
	}
	return userID, true, nil
}

var rotateSessionLua = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then
    return 0
end
redis.call('DEL', KEYS[1])
redis.call('SETEX', KEYS[2], tonumber(ARGV[2]), ARGV[1])
return 1
`)

func (c *RefreshSessionCache) Rotate(ctx context.Context, oldTokenID, newTokenID string, userID int, expiresAt time.Time) (bool, error) {
	ttlSeconds := int64(time.Until(expiresAt).Seconds())
	if ttlSeconds <= 0 {
		return false, nil
	}

	res, err := rotateSessionLua.Run(ctx, c.client,
		[]string{refreshSessionKeyPrefix + oldTokenID, refreshSessionKeyPrefix + newTokenID},
		strconv.Itoa(userID),
		ttlSeconds,
	).Int()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return res == 1, nil
}

func (c *RefreshSessionCache) Delete(ctx context.Context, tokenID string) error {
	return c.client.Del(ctx, refreshSessionKeyPrefix+tokenID).Err()
}
