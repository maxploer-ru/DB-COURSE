package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const commentStatsKey = "comment:stats"

type CommentStatsCache struct {
	client *redis.Client
}

func NewCommentStatsCache(client *redis.Client) *CommentStatsCache {
	return &CommentStatsCache{client: client}
}

func (c *CommentStatsCache) IncrLikes(ctx context.Context, commentID int) error {
	return c.client.HIncrBy(ctx, commentStatsKey, fmt.Sprintf("%d:likes", commentID), 1).Err()
}

func (c *CommentStatsCache) DecrLikes(ctx context.Context, commentID int) error {
	field := fmt.Sprintf("%d:likes", commentID)
	val, err := c.client.HGet(ctx, commentStatsKey, field).Int64()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}
	if val <= 0 {
		return c.client.HSet(ctx, commentStatsKey, field, 0).Err()
	}
	return c.client.HIncrBy(ctx, commentStatsKey, field, -1).Err()
}

func (c *CommentStatsCache) IncrDislikes(ctx context.Context, commentID int) error {
	return c.client.HIncrBy(ctx, commentStatsKey, fmt.Sprintf("%d:dislikes", commentID), 1).Err()
}

func (c *CommentStatsCache) DecrDislikes(ctx context.Context, commentID int) error {
	field := fmt.Sprintf("%d:dislikes", commentID)
	val, err := c.client.HGet(ctx, commentStatsKey, field).Int64()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}
	if val <= 0 {
		return c.client.HSet(ctx, commentStatsKey, field, 0).Err()
	}
	return c.client.HIncrBy(ctx, commentStatsKey, field, -1).Err()
}

func (c *CommentStatsCache) GetStats(ctx context.Context, commentID int) (likes, dislikes int64, hit bool, err error) {
	pipe := c.client.Pipeline()
	likesField := fmt.Sprintf("%d:likes", commentID)
	dislikesField := fmt.Sprintf("%d:dislikes", commentID)
	likesExistsCmd := pipe.HExists(ctx, commentStatsKey, likesField)
	dislikesExistsCmd := pipe.HExists(ctx, commentStatsKey, dislikesField)
	likesCmd := pipe.HGet(ctx, commentStatsKey, likesField)
	dislikesCmd := pipe.HGet(ctx, commentStatsKey, dislikesField)
	_, err = pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, 0, false, err
	}
	if !likesExistsCmd.Val() || !dislikesExistsCmd.Val() {
		return 0, 0, false, nil
	}
	likes, err = strconv.ParseInt(likesCmd.Val(), 10, 64)
	if err != nil {
		return 0, 0, false, fmt.Errorf("parse cached comment likes: %w", err)
	}
	dislikes, err = strconv.ParseInt(dislikesCmd.Val(), 10, 64)
	if err != nil {
		return 0, 0, false, fmt.Errorf("parse cached comment dislikes: %w", err)
	}
	return likes, dislikes, true, nil
}

func (c *CommentStatsCache) SetStats(ctx context.Context, commentID int, likes, dislikes int64) error {
	pipe := c.client.Pipeline()
	pipe.HSet(ctx, commentStatsKey, fmt.Sprintf("%d:likes", commentID), likes)
	pipe.HSet(ctx, commentStatsKey, fmt.Sprintf("%d:dislikes", commentID), dislikes)
	_, err := pipe.Exec(ctx)
	return err
}
