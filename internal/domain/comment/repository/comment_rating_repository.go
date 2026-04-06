package repository

import (
	"ZVideo/internal/domain/comment/entity"
	"context"
)

type CommentRatingRepository interface {
	Create(ctx context.Context, rating *entity.CommentRating) error
	Update(ctx context.Context, rating *entity.CommentRating) error
	Delete(ctx context.Context, userID, commentID int) error
	GetByUserAndComment(ctx context.Context, userID, commentID int) (*entity.CommentRating, error)
	GetCommentRatingStats(ctx context.Context, commentID int) (likes, dislikes int, err error)
}
