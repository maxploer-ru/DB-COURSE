package repositories

import (
	"ZVideo/internal/domain/playlist/entity"
	"context"
	"errors"

	"gorm.io/gorm"
)

type PlaylistItemRepository struct {
	db *gorm.DB
}

func NewPlaylistItemRepository(db *gorm.DB) *PlaylistItemRepository {
	return &PlaylistItemRepository{
		db: db,
	}
}

func (r *PlaylistItemRepository) AddItem(ctx context.Context, item *entity.PlaylistItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *PlaylistItemRepository) RemoveItem(ctx context.Context, playlistID, videoID int) error {
	return r.db.WithContext(ctx).
		Where("playlist_id = ? AND video_id = ?", playlistID, videoID).
		Delete(&entity.PlaylistItem{}).Error
}

func (r *PlaylistItemRepository) GetItemsByPlaylist(ctx context.Context, playlistID int) ([]*entity.PlaylistItem, error) {
	var items []*entity.PlaylistItem
	err := r.db.WithContext(ctx).
		Where("playlist_id = ?", playlistID).
		Order("position ASC").
		Find(&items).Error
	return items, err
}

func (r *PlaylistItemRepository) GetItem(ctx context.Context, playlistID, videoID int) (*entity.PlaylistItem, error) {
	var item entity.PlaylistItem
	err := r.db.WithContext(ctx).
		Where("playlist_id = ? AND video_id = ?", playlistID, videoID).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *PlaylistItemRepository) ReorderItems(ctx context.Context, playlistID int, videoIDs []int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, videoID := range videoIDs {
			if err := tx.Model(&entity.PlaylistItem{}).
				Where("playlist_id = ? AND video_id = ?", playlistID, videoID).
				Update("position", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PlaylistItemRepository) GetMaxPosition(ctx context.Context, playlistID int) (int, error) {
	var maxPos int
	err := r.db.WithContext(ctx).Model(&entity.PlaylistItem{}).
		Where("playlist_id = ?", playlistID).
		Select("COALESCE(MAX(position), 0)").Scan(&maxPos).Error
	return maxPos, err
}
