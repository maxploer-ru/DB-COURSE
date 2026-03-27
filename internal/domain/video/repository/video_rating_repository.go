package repository

import (
	"ZVideo/internal/domain/video/entity"
	"context"
)

type VideoRatingRepository interface {
	Create(ctx context.Context, rating *entity.VideoRating) error
	Update(ctx context.Context, rating *entity.VideoRating) error
	Delete(ctx context.Context, userID, videoID int) error
	GetByUserAndVideo(ctx context.Context, userID, videoID int) (*entity.VideoRating, error)
}
