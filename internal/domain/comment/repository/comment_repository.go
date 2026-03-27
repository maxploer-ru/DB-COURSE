package repository

import (
	"ZVideo/internal/domain/comment/entity"
	"context"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *entity.Comment) error
	GetByID(ctx context.Context, id int) (*entity.Comment, error)
	GetByVideoID(ctx context.Context, videoID int, limit, offset int) ([]*entity.Comment, error)
	Update(ctx context.Context, comment *entity.Comment) error
	Delete(ctx context.Context, id int) error
	GetCountByVideo(ctx context.Context, videoID int) (int, error)
}
