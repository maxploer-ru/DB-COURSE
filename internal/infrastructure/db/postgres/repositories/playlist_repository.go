package repositories

import (
	"ZVideo/internal/domain/playlist/entity"
	"context"
	"errors"

	"gorm.io/gorm"
)

type PlaylistRepository struct {
	db *gorm.DB
}

func NewPlaylistRepository(db *gorm.DB) *PlaylistRepository {
	return &PlaylistRepository{
		db: db,
	}
}

func (r *PlaylistRepository) Create(ctx context.Context, playlist *entity.Playlist) error {
	return r.db.WithContext(ctx).Create(playlist).Error
}

func (r *PlaylistRepository) GetByID(ctx context.Context, id int) (*entity.Playlist, error) {
	var playlist entity.Playlist
	err := r.db.WithContext(ctx).First(&playlist, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &playlist, nil
}

func (r *PlaylistRepository) GetByChannelID(ctx context.Context, channelID int, limit, offset int) ([]*entity.Playlist, error) {
	var playlists []*entity.Playlist
	err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&playlists).Error
	return playlists, err
}

func (r *PlaylistRepository) Update(ctx context.Context, playlist *entity.Playlist) error {
	return r.db.WithContext(ctx).Save(playlist).Error
}

func (r *PlaylistRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&entity.Playlist{}, id).Error
}
