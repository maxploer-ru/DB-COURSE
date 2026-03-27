package repository

import (
	"ZVideo/internal/domain/playlist/entity"
	"context"
)

type PlaylistRepository interface {
	Create(ctx context.Context, playlist *entity.Playlist) error
	GetByID(ctx context.Context, id int) (*entity.Playlist, error)
	GetByChannelID(ctx context.Context, channelID int, limit, offset int) ([]*entity.Playlist, error)
	Update(ctx context.Context, playlist *entity.Playlist) error
	Delete(ctx context.Context, id int) error
}
