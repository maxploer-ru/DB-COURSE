package cache

import (
	"context"
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
		return fmt.Errorf("refresh session ttl is not positive")
	}
	return c.client.Set(ctx, refreshSessionKeyPrefix+tokenID, strconv.Itoa(userID), ttl).Err()
}

func (c *RefreshSessionCache) GetUserID(ctx context.Context, tokenID string) (int, bool, error) {
	val, err := c.client.Get(ctx, refreshSessionKeyPrefix+tokenID).Result()
	if err == redis.Nil {
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

func (c *RefreshSessionCache) Rotate(ctx context.Context, oldTokenID, newTokenID string, userID int, expiresAt time.Time) (bool, error) {
	ttlSeconds := int64(time.Until(expiresAt).Seconds())
	if ttlSeconds <= 0 {
		return false, nil
	}
	const lua = `
local oldKey = KEYS[1]
local newKey = KEYS[2]
if redis.call('EXISTS', oldKey) == 0 then
  return 0
end
redis.call('DEL', oldKey)
redis.call('SETEX', newKey, ARGV[2], ARGV[1])
return 1
`
	res, err := c.client.Eval(ctx, lua,
		[]string{refreshSessionKeyPrefix + oldTokenID, refreshSessionKeyPrefix + newTokenID},
		strconv.Itoa(userID), ttlSeconds,
	).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

func (c *RefreshSessionCache) Delete(ctx context.Context, tokenID string) error {
	return c.client.Del(ctx, refreshSessionKeyPrefix+tokenID).Err()
}
