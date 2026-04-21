package repository

import (
	"ZVideo/internal/domain"
	"context"
)

type PlaylistRepository interface {
	Create(ctx context.Context, playlist *domain.Playlist) error
	GetByID(ctx context.Context, playlistID int) (*domain.Playlist, error)
	ListByChannel(ctx context.Context, channelID int, limit, offset int) ([]*domain.Playlist, error)
	Update(ctx context.Context, playlist *domain.Playlist) error
	Delete(ctx context.Context, playlistID int) error
	AddVideo(ctx context.Context, playlistID, videoID int) error
	RemoveVideo(ctx context.Context, playlistID, videoID int) error
}
