package repository

import (
	"ZVideo/internal/domain/video/entity"
	"context"
)

type VideoRepository interface {
	Create(ctx context.Context, video *entity.Video) error
	GetByID(ctx context.Context, id int) (*entity.Video, error)
	Update(ctx context.Context, video *entity.Video) error
	Delete(ctx context.Context, id int) error

	GetByChannelID(ctx context.Context, channelID int, limit, offset int) ([]*entity.Video, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entity.Video, error)
}
