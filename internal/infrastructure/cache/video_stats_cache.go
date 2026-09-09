package cache

import (
	"ZVideo/internal/domain"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const videoStatsTTL = 1 * time.Hour

type VideoStatsCache struct {
	client *redis.Client
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
	key := videoKey(videoID)
	data, err := c.client.HGetAll(ctx, key).Result()
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
	key := videoKey(videoID)
	pipe := c.client.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"views":    stats.Views,
		"likes":    stats.Likes,
		"dislikes": stats.Dislikes,
		"comments": stats.Comments,
	})
	pipe.Expire(ctx, key, videoStatsTTL)
	_, err := pipe.Exec(ctx)
	return err
}
