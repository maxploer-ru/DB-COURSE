package repository

import (
	"ZVideo/internal/domain"
	"context"
)

type VideoRepository interface {
	Create(ctx context.Context, video *domain.Video) error
	GetByID(ctx context.Context, id int) (*domain.Video, error)
	Update(ctx context.Context, video *domain.Video) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, offset, limit int) ([]*domain.Video, error)
	ListByChannel(ctx context.Context, channelID int, limit, offset int) ([]*domain.Video, error)
	ListFilepathsByChannel(ctx context.Context, channelID int) ([]string, error)
}
