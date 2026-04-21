package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const subscriberCountKey = "channel:subscribers"

type SubscriberCounter struct {
	client *redis.Client
}

func NewRedisSubscriberCounter(client *redis.Client) *SubscriberCounter {
	return &SubscriberCounter{client: client}
}

func (r *SubscriberCounter) Increment(ctx context.Context, channelID int) error {
	return r.client.HIncrBy(ctx, subscriberCountKey, strconv.Itoa(channelID), 1).Err()
}

func (r *SubscriberCounter) Decrement(ctx context.Context, channelID int) error {
	field := strconv.Itoa(channelID)
	val, err := r.client.HGet(ctx, subscriberCountKey, field).Int64()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}
	if val <= 0 {
		return r.client.HSet(ctx, subscriberCountKey, field, 0).Err()
	}
	return r.client.HIncrBy(ctx, subscriberCountKey, field, -1).Err()
}

func (r *SubscriberCounter) Get(ctx context.Context, channelID int) (int, bool, error) {
	val, err := r.client.HGet(ctx, subscriberCountKey, strconv.Itoa(channelID)).Result()
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
	data, err := r.client.HGetAll(ctx, subscriberCountKey).Result()
	if err != nil {
		return nil, err
	}
	result := make(map[int]int, len(data))
	for k, v := range data {
		id, _ := strconv.Atoi(k)
		cnt, _ := strconv.Atoi(v)
		result[id] = cnt
	}
	return result, nil
}

func (r *SubscriberCounter) Set(ctx context.Context, channelID int, count int) error {
	return r.client.HSet(ctx, subscriberCountKey, strconv.Itoa(channelID), count).Err()
}
