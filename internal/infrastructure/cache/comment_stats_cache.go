package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const commentStatsTTL = 1 * time.Hour

type CommentStatsCache struct {
	client *redis.Client
}

func NewCommentStatsCache(client *redis.Client) *CommentStatsCache {
	return &CommentStatsCache{client: client}
}

func commentKey(commentID int) string {
	return fmt.Sprintf("comment:stats:%d", commentID)
}

func (c *CommentStatsCache) incrField(ctx context.Context, commentID int, field string) error {
	key := commentKey(commentID)
	pipe := c.client.Pipeline()
	pipe.HIncrBy(ctx, key, field, 1)
	pipe.Expire(ctx, key, commentStatsTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *CommentStatsCache) IncrLikes(ctx context.Context, commentID int) error {
	return c.incrField(ctx, commentID, "likes")
}

func (c *CommentStatsCache) IncrDislikes(ctx context.Context, commentID int) error {
	return c.incrField(ctx, commentID, "dislikes")
}

var decrCommentLua = redis.NewScript(`
local current = redis.call('HGET', KEYS[1], ARGV[1])
if current and tonumber(current) > 0 then
    return redis.call('HINCRBY', KEYS[1], ARGV[1], -1)
end
return 0
`)

func (c *CommentStatsCache) decrField(ctx context.Context, commentID int, field string) error {
	return decrCommentLua.Run(ctx, c.client, []string{commentKey(commentID)}, field).Err()
}

func (c *CommentStatsCache) DecrLikes(ctx context.Context, commentID int) error {
	return c.decrField(ctx, commentID, "likes")
}

func (c *CommentStatsCache) DecrDislikes(ctx context.Context, commentID int) error {
	return c.decrField(ctx, commentID, "dislikes")
}

func (c *CommentStatsCache) GetStats(ctx context.Context, commentID int) (likes, dislikes int64, hit bool, err error) {
	key := commentKey(commentID)

	data, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return 0, 0, false, err
	}
	if len(data) == 0 {
		return 0, 0, false, nil
	}

	likes, _ = strconv.ParseInt(data["likes"], 10, 64)
	dislikes, _ = strconv.ParseInt(data["dislikes"], 10, 64)

	return likes, dislikes, true, nil
}

func (c *CommentStatsCache) SetStats(ctx context.Context, commentID int, likes, dislikes int64) error {
	key := commentKey(commentID)
	pipe := c.client.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"likes":    likes,
		"dislikes": dislikes,
	})
	pipe.Expire(ctx, key, commentStatsTTL)
	_, err := pipe.Exec(ctx)
	return err
}
