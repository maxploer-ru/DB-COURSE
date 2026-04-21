package repository

import (
	"ZVideo/internal/domain"
	"context"
)

type VideoStatsCache interface {
	IncrViews(ctx context.Context, videoID int) error

	IncrLikes(ctx context.Context, videoID int) error
	DecrLikes(ctx context.Context, videoID int) error
	IncrDislikes(ctx context.Context, videoID int) error
	DecrDislikes(ctx context.Context, videoID int) error

	IncrComments(ctx context.Context, videoID int) error
	DecrComments(ctx context.Context, videoID int) error
	GetCommentsCount(ctx context.Context, videoID int) (count int64, hit bool, err error)
	SetCommentsCount(ctx context.Context, videoID int, count int64) error

	GetStats(ctx context.Context, videoID int) (stats *domain.VideoStats, hit bool, err error)
	LoadAll(ctx context.Context) (map[int]domain.VideoStats, error)
	SetStats(ctx context.Context, videoID int, stats *domain.VideoStats) error
}
