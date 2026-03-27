package repository

import (
	"ZVideo/internal/domain/video/entity"
	"context"
)

type VideoStatsCache interface {
	Get(ctx context.Context, videoID int) (*entity.VideoStats, bool, error)
	Set(ctx context.Context, videoID int, stats *entity.VideoStats) error
	Invalidate(ctx context.Context, videoID int) error
}
