package cache

import (
	"ZVideo/internal/domain"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const videoStatsKey = "video:stats"

type VideoStatsCache struct {
	client *redis.Client
}

func NewVideoStatsCache(client *redis.Client) *VideoStatsCache {
	return &VideoStatsCache{client: client}
}

func (c *VideoStatsCache) IncrViews(ctx context.Context, videoID int) error {
	return c.client.HIncrBy(ctx, videoStatsKey, fmt.Sprintf("%d:views", videoID), 1).Err()
}

func (c *VideoStatsCache) IncrLikes(ctx context.Context, videoID int) error {
	return c.client.HIncrBy(ctx, videoStatsKey, fmt.Sprintf("%d:likes", videoID), 1).Err()
}

func (c *VideoStatsCache) DecrLikes(ctx context.Context, videoID int) error {
	field := fmt.Sprintf("%d:likes", videoID)
	val, err := c.client.HGet(ctx, videoStatsKey, field).Int64()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}
	if val <= 0 {
		return c.client.HSet(ctx, videoStatsKey, field, 0).Err()
	}
	return c.client.HIncrBy(ctx, videoStatsKey, field, -1).Err()
}

func (c *VideoStatsCache) IncrDislikes(ctx context.Context, videoID int) error {
	return c.client.HIncrBy(ctx, videoStatsKey, fmt.Sprintf("%d:dislikes", videoID), 1).Err()
}

func (c *VideoStatsCache) DecrDislikes(ctx context.Context, videoID int) error {
	field := fmt.Sprintf("%d:dislikes", videoID)
	val, err := c.client.HGet(ctx, videoStatsKey, field).Int64()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}
	if val <= 0 {
		return c.client.HSet(ctx, videoStatsKey, field, 0).Err()
	}
	return c.client.HIncrBy(ctx, videoStatsKey, field, -1).Err()
}

func (c *VideoStatsCache) IncrComments(ctx context.Context, videoID int) error {
	return c.client.HIncrBy(ctx, videoStatsKey, fmt.Sprintf("%d:comments", videoID), 1).Err()
}

func (c *VideoStatsCache) DecrComments(ctx context.Context, videoID int) error {
	field := fmt.Sprintf("%d:comments", videoID)
	val, err := c.client.HGet(ctx, videoStatsKey, field).Int64()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}
	if val <= 0 {
		return c.client.HSet(ctx, videoStatsKey, field, 0).Err()
	}
	return c.client.HIncrBy(ctx, videoStatsKey, field, -1).Err()
}

func (c *VideoStatsCache) GetCommentsCount(ctx context.Context, videoID int) (int64, bool, error) {
	val, err := c.client.HGet(ctx, videoStatsKey, fmt.Sprintf("%d:comments", videoID)).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return val, true, nil
}

func (c *VideoStatsCache) SetCommentsCount(ctx context.Context, videoID int, count int64) error {
	return c.client.HSet(ctx, videoStatsKey, fmt.Sprintf("%d:comments", videoID), count).Err()
}

func (c *VideoStatsCache) GetStats(ctx context.Context, videoID int) (stats *domain.VideoStats, hit bool, err error) {
	pipe := c.client.Pipeline()
	viewsField := fmt.Sprintf("%d:views", videoID)
	likesField := fmt.Sprintf("%d:likes", videoID)
	dislikesField := fmt.Sprintf("%d:dislikes", videoID)
	commentsField := fmt.Sprintf("%d:comments", videoID)
	viewsExistsCmd := pipe.HExists(ctx, videoStatsKey, viewsField)
	likesExistsCmd := pipe.HExists(ctx, videoStatsKey, likesField)
	dislikesExistsCmd := pipe.HExists(ctx, videoStatsKey, dislikesField)
	commentsExistsCmd := pipe.HExists(ctx, videoStatsKey, commentsField)
	viewsCmd := pipe.HGet(ctx, videoStatsKey, viewsField)
	likesCmd := pipe.HGet(ctx, videoStatsKey, likesField)
	dislikesCmd := pipe.HGet(ctx, videoStatsKey, dislikesField)
	commentsCmd := pipe.HGet(ctx, videoStatsKey, commentsField)
	_, err = pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return &domain.VideoStats{}, false, err
	}
	if !viewsExistsCmd.Val() || !likesExistsCmd.Val() || !dislikesExistsCmd.Val() || !commentsExistsCmd.Val() {
		return &domain.VideoStats{}, false, nil
	}

	views, err := strconv.Atoi(viewsCmd.Val())
	if err != nil {
		return &domain.VideoStats{}, false, fmt.Errorf("parse cached video views: %w", err)
	}
	likes, err := strconv.Atoi(likesCmd.Val())
	if err != nil {
		return &domain.VideoStats{}, false, fmt.Errorf("parse cached video likes: %w", err)
	}
	dislikes, err := strconv.Atoi(dislikesCmd.Val())
	if err != nil {
		return &domain.VideoStats{}, false, fmt.Errorf("parse cached video dislikes: %w", err)
	}
	comments, err := strconv.Atoi(commentsCmd.Val())
	if err != nil {
		return &domain.VideoStats{}, false, fmt.Errorf("parse cached video comments: %w", err)
	}
	return &domain.VideoStats{
		Views:    views,
		Likes:    likes,
		Dislikes: dislikes,
		Comments: comments,
	}, true, nil
}

func (c *VideoStatsCache) LoadAll(ctx context.Context) (map[int]domain.VideoStats, error) {
	data, err := c.client.HGetAll(ctx, videoStatsKey).Result()
	if err != nil {
		return nil, err
	}
	result := make(map[int]domain.VideoStats)
	for key, val := range data {
		var videoID int
		var field string
		if _, err := fmt.Sscanf(key, "%d:%s", &videoID, &field); err != nil {
			continue
		}
		stats := result[videoID]
		switch field {
		case "views":
			stats.Views, _ = strconv.Atoi(val)
		case "likes":
			stats.Likes, _ = strconv.Atoi(val)
		case "dislikes":
			stats.Dislikes, _ = strconv.Atoi(val)
		}
		result[videoID] = stats
	}
	return result, nil
}

func (c *VideoStatsCache) SetStats(ctx context.Context, videoID int, stats *domain.VideoStats) error {
	pipe := c.client.Pipeline()
	pipe.HSet(ctx, videoStatsKey, fmt.Sprintf("%d:views", videoID), stats.Views)
	pipe.HSet(ctx, videoStatsKey, fmt.Sprintf("%d:likes", videoID), stats.Likes)
	pipe.HSet(ctx, videoStatsKey, fmt.Sprintf("%d:dislikes", videoID), stats.Dislikes)
	pipe.HSet(ctx, videoStatsKey, fmt.Sprintf("%d:comments", videoID), stats.Comments)
	_, err := pipe.Exec(ctx)
	return err
}
