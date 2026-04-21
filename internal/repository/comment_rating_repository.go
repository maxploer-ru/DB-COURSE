package repository

import (
	"ZVideo/internal/domain"
	"context"
)

type CommentRatingRepository interface {
	Create(ctx context.Context, rating *domain.CommentRating) error
	Update(ctx context.Context, rating *domain.CommentRating) error
	Delete(ctx context.Context, userID, commentID int) error
	GetByUserAndComment(ctx context.Context, userID, commentID int) (*domain.CommentRating, error)
	GetStats(ctx context.Context, commentID int) (likes, dislikes int64, err error)
}
