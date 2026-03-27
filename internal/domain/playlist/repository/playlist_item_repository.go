package repository

import (
	"ZVideo/internal/domain/playlist/entity"
	"context"
)

type PlaylistItemRepository interface {
	AddItem(ctx context.Context, item *entity.PlaylistItem) error
	RemoveItem(ctx context.Context, playlistID, videoID int) error
	GetItemsByPlaylist(ctx context.Context, playlistID int) ([]*entity.PlaylistItem, error)
	GetItem(ctx context.Context, playlistID, videoID int) (*entity.PlaylistItem, error)
	ReorderItems(ctx context.Context, playlistID int, videoIDs []int) error
	GetMaxPosition(ctx context.Context, playlistID int) (int, error)
}
