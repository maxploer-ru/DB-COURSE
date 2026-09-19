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
	return mutateCommentStats(ctx, c.client, key, field, 1)
}

func (c *CommentStatsCache) IncrLikes(ctx context.Context, commentID int) error {
	return c.incrField(ctx, commentID, "likes")
}

func (c *CommentStatsCache) IncrDislikes(ctx context.Context, commentID int) error {
	return c.incrField(ctx, commentID, "dislikes")
}

var mutateCommentStatsLua = redis.NewScript(`
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

func mutateCommentStats(ctx context.Context, client *redis.Client, key, field string, delta int64) error {
	redisCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	return mutateCommentStatsLua.Run(redisCtx, client, []string{key}, field, delta, int(commentStatsTTL/time.Second)).Err()
}

func (c *CommentStatsCache) decrField(ctx context.Context, commentID int, field string) error {
	return mutateCommentStats(ctx, c.client, commentKey(commentID), field, -1)
}

func (c *CommentStatsCache) DecrLikes(ctx context.Context, commentID int) error {
	return c.decrField(ctx, commentID, "likes")
}

func (c *CommentStatsCache) DecrDislikes(ctx context.Context, commentID int) error {
	return c.decrField(ctx, commentID, "dislikes")
}

func (c *CommentStatsCache) GetStats(ctx context.Context, commentID int) (likes, dislikes int64, hit bool, err error) {
	redisCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	key := commentKey(commentID)

	data, err := c.client.HGetAll(redisCtx, key).Result()
	if err != nil {
		return 0, 0, false, err
	}
	if data["complete"] != "1" {
		return 0, 0, false, nil
	}

	likes, err = strconv.ParseInt(data["likes"], 10, 64)
	if err != nil {
		return 0, 0, false, fmt.Errorf("parse cached likes for comment %d: %w", commentID, err)
	}
	dislikes, err = strconv.ParseInt(data["dislikes"], 10, 64)
	if err != nil {
		return 0, 0, false, fmt.Errorf("parse cached dislikes for comment %d: %w", commentID, err)
	}

	return likes, dislikes, true, nil
}

func (c *CommentStatsCache) SetStats(ctx context.Context, commentID int, likes, dislikes int64) error {
	redisCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	key := commentKey(commentID)
	pipe := c.client.TxPipeline()
	pipe.HSet(redisCtx, key, map[string]any{
		"likes":    likes,
		"dislikes": dislikes,
		"complete": 1,
	})
	pipe.Expire(redisCtx, key, commentStatsTTL)
	_, err := pipe.Exec(redisCtx)
	return err
}
