package repository

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/mappers"
	"ZVideo/internal/infrastructure/db/postgres/models"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type PlaylistRepository struct {
	db *gorm.DB
}

func NewPlaylistRepository(db *gorm.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

func (r *PlaylistRepository) Create(ctx context.Context, playlist *domain.Playlist) error {
	model := mappers.FromDomainPlaylist(playlist)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("create playlist failed: %w", err)
	}
	playlist.ID = model.ID
	return nil
}

func (r *PlaylistRepository) GetByID(ctx context.Context, playlistID int) (*domain.Playlist, error) {
	var model models.Playlist
	err := r.db.WithContext(ctx).
		Preload("PlaylistVideos", func(db *gorm.DB) *gorm.DB {
			return db.Order("number ASC")
		}).
		Preload("PlaylistVideos.Video").
		First(&model, playlistID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get playlist failed: %w", err)
	}
	return mappers.ToDomainPlaylist(&model), nil
}

func (r *PlaylistRepository) ListByChannel(ctx context.Context, channelID int, limit, offset int) ([]*domain.Playlist, error) {
	var modelsList []models.Playlist
	err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("PlaylistVideos", func(db *gorm.DB) *gorm.DB {
			return db.Order("number ASC")
		}).
		Preload("PlaylistVideos.Video").
		Find(&modelsList).Error
	if err != nil {
		return nil, fmt.Errorf("list playlists by channel failed: %w", err)
	}
	return mappers.ToDomainPlaylistList(modelsList), nil
}

func (r *PlaylistRepository) Update(ctx context.Context, playlist *domain.Playlist) error {
	result := r.db.WithContext(ctx).
		Model(&models.Playlist{}).
		Where("id = ?", playlist.ID).
		Updates(map[string]interface{}{
			"name":        playlist.Name,
			"description": playlist.Description,
		})
	if result.Error != nil {
		return fmt.Errorf("update playlist failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrPlaylistNotFound
	}
	return nil
}

func (r *PlaylistRepository) Delete(ctx context.Context, playlistID int) error {
	res := r.db.WithContext(ctx).Delete(&models.Playlist{}, playlistID)
	if res.Error != nil {
		return fmt.Errorf("delete playlist failed: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrPlaylistNotFound
	}
	return nil
}

func (r *PlaylistRepository) AddVideo(ctx context.Context, playlistID, videoID int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.PlaylistItem{}).
			Where("playlist_id = ? AND video_id = ?", playlistID, videoID).
			Count(&count).Error; err != nil {
			return fmt.Errorf("check playlist item failed: %w", err)
		}
		if count > 0 {
			return nil
		}

		var maxNumber int
		if err := tx.Model(&models.PlaylistItem{}).
			Where("playlist_id = ?", playlistID).
			Select("COALESCE(MAX(number), 0)").
			Scan(&maxNumber).Error; err != nil {
			return fmt.Errorf("get playlist max number failed: %w", err)
		}

		item := &models.PlaylistItem{
			PlaylistID: playlistID,
			VideoID:    videoID,
			Number:     maxNumber + 1,
		}
		if err := tx.Create(item).Error; err != nil {
			return fmt.Errorf("add video to playlist failed: %w", err)
		}
		return nil
	})
}

func (r *PlaylistRepository) RemoveVideo(ctx context.Context, playlistID, videoID int) error {
	res := r.db.WithContext(ctx).
		Where("playlist_id = ? AND video_id = ?", playlistID, videoID).
		Delete(&models.PlaylistItem{})
	if res.Error != nil {
		return fmt.Errorf("remove video from playlist failed: %w", res.Error)
	}
	return nil
}
