package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const subscriberCountTTL = 1 * time.Hour

type SubscriberCounter struct {
	client *redis.Client
}

func NewRedisSubscriberCounter(client *redis.Client) *SubscriberCounter {
	return &SubscriberCounter{client: client}
}

func subKey(channelID int) string {
	return fmt.Sprintf("channel:subscribers:%d", channelID)
}

func (r *SubscriberCounter) Increment(ctx context.Context, channelID int) error {
	return incrementExistingSubscriberCount.Run(ctx, r.client, []string{subKey(channelID)}, int(subscriberCountTTL/time.Second)).Err()
}

var incrementExistingSubscriberCount = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then return 0 end
redis.call('INCR', KEYS[1])
redis.call('EXPIRE', KEYS[1], ARGV[1])
return 1
`)

var decrSubLua = redis.NewScript(`
local current = redis.call('GET', KEYS[1])
if current and tonumber(current) > 0 then
    redis.call('DECR', KEYS[1])
    redis.call('EXPIRE', KEYS[1], ARGV[1])
    return 1
end
return 0
`)

func (r *SubscriberCounter) Decrement(ctx context.Context, channelID int) error {
	key := subKey(channelID)
	return decrSubLua.Run(ctx, r.client, []string{key}, int(subscriberCountTTL/time.Second)).Err()
}

func (r *SubscriberCounter) Get(ctx context.Context, channelID int) (int, bool, error) {
	redisCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	val, err := r.client.Get(redisCtx, subKey(channelID)).Result()
	if errors.Is(err, redis.Nil) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	count, convErr := strconv.Atoi(val)
	if convErr != nil {
		return 0, false, fmt.Errorf("parse cached subscriber count: %w", convErr)
	}
	return count, true, nil
}

func (r *SubscriberCounter) LoadAll(ctx context.Context) (map[int]int, error) {
	result := make(map[int]int)
	var cursor uint64
	match := "channel:subscribers:*"

	for {
		var keys []string
		var err error

		keys, cursor, err = r.client.Scan(ctx, cursor, match, 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			val, err := r.client.Get(ctx, key).Result()
			if errors.Is(err, redis.Nil) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("load subscriber count %q: %w", key, err)
			}

			parts := strings.Split(key, ":")
			if len(parts) != 3 {
				return nil, fmt.Errorf("invalid subscriber cache key %q", key)
			}
			id, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, fmt.Errorf("parse subscriber cache key %q: %w", key, err)
			}
			if id <= 0 {
				return nil, fmt.Errorf("subscriber cache key %q contains invalid channel id", key)
			}
			count, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("parse subscriber count for channel %d: %w", id, err)
			}
			result[id] = count
		}

		if cursor == 0 {
			break
		}
	}
	return result, nil
}

func (r *SubscriberCounter) Set(ctx context.Context, channelID int, count int) error {
	redisCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	key := subKey(channelID)

	return r.client.Set(redisCtx, key, count, subscriberCountTTL).Err()
}
