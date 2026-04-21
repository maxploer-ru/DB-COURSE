package repository

import (
	"ZVideo/internal/domain"
	"context"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *domain.Comment) error
	GetByID(ctx context.Context, id int) (*domain.Comment, error)
	ListByVideo(ctx context.Context, videoID int, limit, offset int) ([]*domain.Comment, error)
	Update(ctx context.Context, comment *domain.Comment) error
	Delete(ctx context.Context, id int) error
	CountByVideo(ctx context.Context, videoID int) (int64, error)
}
