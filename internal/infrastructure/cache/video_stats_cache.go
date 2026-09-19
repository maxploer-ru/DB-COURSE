package cache

import (
	"ZVideo/internal/domain"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const videoStatsTTL = 1 * time.Hour

type VideoStatsCache struct {
	client *redis.Client
}

func (c *VideoStatsCache) GetCommentsCount(ctx context.Context, videoID int) (count int64, hit bool, err error) {
	redisCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	data, err := c.client.HGetAll(redisCtx, videoKey(videoID)).Result()
	if errors.Is(err, redis.Nil) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	if data["complete"] != "1" && data["mode"] != "comments_only" {
		return 0, false, nil
	}
	value, ok := data["comments"]
	if !ok {
		return 0, false, nil
	}

	count, err = strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("parse cached comments count: %w", err)
	}
	return count, true, nil
}

func (c *VideoStatsCache) SetCommentsCount(ctx context.Context, videoID int, count int64) error {
	redisCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	key := videoKey(videoID)
	pipe := c.client.TxPipeline()
	pipe.HSet(redisCtx, key, map[string]any{"comments": count, "mode": "comments_only"})
	pipe.HDel(redisCtx, key, "complete")
	pipe.Expire(redisCtx, key, videoStatsTTL)
	_, err := pipe.Exec(redisCtx)
	return err
}

func (c *VideoStatsCache) LoadAll(ctx context.Context) (map[int]domain.VideoStats, error) {
	result := make(map[int]domain.VideoStats)
	var cursor uint64

	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, "video:stats:*", 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			id, err := strconv.Atoi(strings.TrimPrefix(key, "video:stats:"))
			if err != nil {
				return nil, fmt.Errorf("parse video stats key %q: %w", key, err)
			}

			data, err := c.client.HGetAll(ctx, key).Result()
			if err != nil {
				return nil, err
			}
			if len(data) == 0 {
				continue
			}
			if data["complete"] != "1" {
				continue
			}

			stats := domain.VideoStats{}
			values := map[string]*int{
				"views":    &stats.Views,
				"likes":    &stats.Likes,
				"dislikes": &stats.Dislikes,
				"comments": &stats.Comments,
			}
			for field, target := range values {
				value, ok := data[field]
				if !ok {
					return nil, fmt.Errorf("cached video %d is missing %s", id, field)
				}
				parsed, err := strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("parse cached %s for video %d: %w", field, id, err)
				}
				*target = parsed
			}
			result[id] = stats
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return result, nil
}

func NewVideoStatsCache(client *redis.Client) *VideoStatsCache {
	return &VideoStatsCache{client: client}
}

func videoKey(videoID int) string {
	return fmt.Sprintf("video:stats:%d", videoID)
}

func (c *VideoStatsCache) incrField(ctx context.Context, videoID int, field string) error {
	key := videoKey(videoID)
	return mutateVideoStats(ctx, c.client, key, field, 1)
}

func (c *VideoStatsCache) IncrViews(ctx context.Context, videoID int) error {
	return c.incrField(ctx, videoID, "views")
}

func (c *VideoStatsCache) IncrLikes(ctx context.Context, videoID int) error {
	return c.incrField(ctx, videoID, "likes")
}

func (c *VideoStatsCache) IncrDislikes(ctx context.Context, videoID int) error {
	return c.incrField(ctx, videoID, "dislikes")
}

func (c *VideoStatsCache) IncrComments(ctx context.Context, videoID int) error {
	return c.incrField(ctx, videoID, "comments")
}

var mutateVideoStatsLua = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'complete') ~= '1' then
    redis.call('DEL', KEYS[1])
    return 0
end
local current = redis.call('HGET', KEYS[1], ARGV[1])
local delta = tonumber(ARGV[2])
if delta < 0 and (not current or tonumber(current) <= 0) then
    return 0
end
redis.call('HINCRBY', KEYS[1], ARGV[1], delta)
redis.call('EXPIRE', KEYS[1], ARGV[3])
return 1
`)

func mutateVideoStats(ctx context.Context, client *redis.Client, key, field string, delta int64) error {
	redisCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	return mutateVideoStatsLua.Run(redisCtx, client, []string{key}, field, delta, int(videoStatsTTL/time.Second)).Err()
}

func (c *VideoStatsCache) decrField(ctx context.Context, videoID int, field string) error {
	return mutateVideoStats(ctx, c.client, videoKey(videoID), field, -1)
}

func (c *VideoStatsCache) DecrLikes(ctx context.Context, videoID int) error {
	return c.decrField(ctx, videoID, "likes")
}

func (c *VideoStatsCache) DecrDislikes(ctx context.Context, videoID int) error {
	return c.decrField(ctx, videoID, "dislikes")
}

func (c *VideoStatsCache) DecrComments(ctx context.Context, videoID int) error {
	return c.decrField(ctx, videoID, "comments")
}

func (c *VideoStatsCache) GetStats(ctx context.Context, videoID int) (*domain.VideoStats, bool, error) {
	redisCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	key := videoKey(videoID)
	data, err := c.client.HGetAll(redisCtx, key).Result()
	if err != nil {
		return nil, false, err
	}
	if data["complete"] != "1" {
		return nil, false, nil
	}
	var views, likes, dislikes, comments int

	values := []struct {
		name string
		to   *int
	}{
		{"views", &views},
		{"likes", &likes},
		{"dislikes", &dislikes},
		{"comments", &comments},
	}
	for _, value := range values {
		parsed, parseErr := strconv.Atoi(data[value.name])
		if parseErr != nil {
			return nil, false, fmt.Errorf("parse cached %s for video %d: %w", value.name, videoID, parseErr)
		}
		*value.to = parsed
	}

	return &domain.VideoStats{
		Views:    views,
		Likes:    likes,
		Dislikes: dislikes,
		Comments: comments,
	}, true, nil
}

func (c *VideoStatsCache) SetStats(ctx context.Context, videoID int, stats *domain.VideoStats) error {
	redisCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	key := videoKey(videoID)
	pipe := c.client.TxPipeline()
	pipe.HSet(redisCtx, key, map[string]any{
		"views":    stats.Views,
		"likes":    stats.Likes,
		"dislikes": stats.Dislikes,
		"comments": stats.Comments,
		"complete": 1,
	})
	pipe.HDel(redisCtx, key, "mode")
	pipe.Expire(redisCtx, key, videoStatsTTL)
	_, err := pipe.Exec(redisCtx)
	return err
}
