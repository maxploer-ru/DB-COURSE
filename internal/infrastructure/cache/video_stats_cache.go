package cache

import (
	"ZVideo/internal/domain"
	"context"
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

	value, err := c.client.HGet(redisCtx, videoKey(videoID), "comments").Result()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
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
	pipe := c.client.Pipeline()
	pipe.HSet(redisCtx, key, "comments", count)
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

			stats := domain.VideoStats{}
			for field, target := range map[string]*int{
				"views":    &stats.Views,
				"likes":    &stats.Likes,
				"dislikes": &stats.Dislikes,
				"comments": &stats.Comments,
			} {
				if value, ok := data[field]; ok {
					parsed, err := strconv.Atoi(value)
					if err != nil {
						return nil, fmt.Errorf("parse cached %s for video %d: %w", field, id, err)
					}
					*target = parsed
				}
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
	pipe := c.client.Pipeline()
	pipe.HIncrBy(ctx, key, field, 1)
	pipe.Expire(ctx, key, videoStatsTTL)
	_, err := pipe.Exec(ctx)
	return err
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

var decrLua = redis.NewScript(`
local current = redis.call('HGET', KEYS[1], ARGV[1])
if current and tonumber(current) > 0 then
    return redis.call('HINCRBY', KEYS[1], ARGV[1], -1)
end
return 0
`)

func (c *VideoStatsCache) decrField(ctx context.Context, videoID int, field string) error {
	return decrLua.Run(ctx, c.client, []string{videoKey(videoID)}, field).Err()
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
	if len(data) == 0 {
		return nil, false, nil
	}

	views, _ := strconv.Atoi(data["views"])
	likes, _ := strconv.Atoi(data["likes"])
	dislikes, _ := strconv.Atoi(data["dislikes"])
	comments, _ := strconv.Atoi(data["comments"])

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
	pipe := c.client.Pipeline()
	pipe.HSet(redisCtx, key, map[string]interface{}{
		"views":    stats.Views,
		"likes":    stats.Likes,
		"dislikes": stats.Dislikes,
		"comments": stats.Comments,
	})
	pipe.Expire(ctx, key, videoStatsTTL)
	_, err := pipe.Exec(ctx)
	return err
}
