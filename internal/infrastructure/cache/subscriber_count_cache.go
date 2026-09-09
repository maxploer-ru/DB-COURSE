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
	key := subKey(channelID)
	pipe := r.client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, subscriberCountTTL)
	_, err := pipe.Exec(ctx)
	return err
}

var decrSubLua = redis.NewScript(`
local current = redis.call('GET', KEYS[1])
if current and tonumber(current) > 0 then
    redis.call('DECR', KEYS[1])
    return 1
end
return 0
`)

func (r *SubscriberCounter) Decrement(ctx context.Context, channelID int) error {
	key := subKey(channelID)
	err := decrSubLua.Run(ctx, r.client, []string{key}).Err()
	if err == nil {
		r.client.Expire(ctx, key, subscriberCountTTL)
	}
	return err
}

func (r *SubscriberCounter) Get(ctx context.Context, channelID int) (int, bool, error) {
	val, err := r.client.Get(ctx, subKey(channelID)).Result()
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
			if err == nil {
				parts := strings.Split(key, ":")
				if len(parts) == 3 {
					id, _ := strconv.Atoi(parts[2])
					cnt, _ := strconv.Atoi(val)
					result[id] = cnt
				}
			}
		}

		if cursor == 0 {
			break
		}
	}
	return result, nil
}

func (r *SubscriberCounter) Set(ctx context.Context, channelID int, count int) error {
	key := subKey(channelID)

	return r.client.Set(ctx, key, count, subscriberCountTTL).Err()
}
