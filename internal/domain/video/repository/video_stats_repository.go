package repository

import "context"

type VideoStatsRepository interface {
	GetViewsCount(ctx context.Context, videoID int) (int, error)
	GetLikesDislikes(ctx context.Context, videoID int) (int, int, error)
	GetCommentsCount(ctx context.Context, videoID int) (int, error)
}
