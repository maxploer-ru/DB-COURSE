package repository

import (
	"ZVideo/internal/domain"
	"context"
)

type VideoRatingRepository interface {
	Create(ctx context.Context, rating *domain.VideoRating) error
	Update(ctx context.Context, rating *domain.VideoRating) error
	Delete(ctx context.Context, userID, videoID int) error
	GetByUserAndVideo(ctx context.Context, userID, videoID int) (*domain.VideoRating, error)
	GetStats(ctx context.Context, videoID int) (likes, dislikes int, err error)
}
