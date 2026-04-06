package redis

import (
	"ZVideo/internal/domain/video/entity"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type VideoStatsCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewVideoStatsCache(client *redis.Client, ttl time.Duration) *VideoStatsCache {
	return &VideoStatsCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *VideoStatsCache) Get(ctx context.Context, videoID int) (*entity.VideoStats, bool, error) {
	key := fmt.Sprintf("video_stats:%d", videoID)
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}

	var stats entity.VideoStats
	if err := json.Unmarshal([]byte(val), &stats); err != nil {
		return nil, false, err
	}

	return &stats, true, nil
}

func (c *VideoStatsCache) Set(ctx context.Context, videoID int, stats *entity.VideoStats) error {
	key := fmt.Sprintf("video_stats:%d", videoID)
	bytes, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, bytes, c.ttl).Err()
}

func (c *VideoStatsCache) Invalidate(ctx context.Context, videoID int) error {
	key := fmt.Sprintf("video_stats:%d", videoID)
	return c.client.Del(ctx, key).Err()
}
