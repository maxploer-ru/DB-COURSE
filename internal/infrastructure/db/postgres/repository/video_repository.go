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

type VideoRepository struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

func (r *VideoRepository) Create(ctx context.Context, video *domain.Video) error {
	dbVideo := mappers.FromDomainVideo(video)

	if err := r.db.WithContext(ctx).Create(dbVideo).Error; err != nil {
		return fmt.Errorf("create video: %w", err)
	}

	video.ID = dbVideo.ID
	return nil
}

func (r *VideoRepository) GetByID(ctx context.Context, id int) (*domain.Video, error) {
	var dbVideo models.Video
	err := r.db.WithContext(ctx).
		Preload("Channel").
		First(&dbVideo, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get video by id: %w", err)
	}
	return mappers.ToDomainVideo(&dbVideo), nil
}

func (r *VideoRepository) Update(ctx context.Context, video *domain.Video) error {
	dbVideo := mappers.FromDomainVideo(video)

	result := r.db.WithContext(ctx).
		Model(&models.Video{}).
		Where("id = ?", dbVideo.ID).
		Updates(map[string]interface{}{
			"title":       dbVideo.Title,
			"description": dbVideo.Description,
			"filepath":    dbVideo.Filepath,
		})

	if result.Error != nil {
		return fmt.Errorf("update video: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrVideoNotFound
	}
	return nil
}

func (r *VideoRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&models.Video{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete video: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrVideoNotFound
	}
	return nil
}

func (r *VideoRepository) List(ctx context.Context, limit, offset int) ([]*domain.Video, error) {
	var dbVideos []models.Video
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Preload("Channel").
		Find(&dbVideos).Error
	if err != nil {
		return nil, fmt.Errorf("list videos: %w", err)
	}
	domainVideos := make([]*domain.Video, 0, len(dbVideos)) // TODO: mapper ToDomainVideoList
	for _, dbVideo := range dbVideos {
		domainVideos = append(domainVideos, mappers.ToDomainVideo(&dbVideo))
	}
	return domainVideos, nil
}

func (r *VideoRepository) ListByChannel(ctx context.Context, channelID int, limit, offset int) ([]*domain.Video, error) {
	var dbVideos []models.Video
	err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Channel").
		Find(&dbVideos).Error

	if err != nil {
		return nil, fmt.Errorf("list videos by channel: %w", err)
	}

	domainVideos := make([]*domain.Video, 0, len(dbVideos)) // TODO: mapper ToDomainVideoList
	for _, dbVideo := range dbVideos {
		domainVideos = append(domainVideos, mappers.ToDomainVideo(&dbVideo))
	}
	return domainVideos, nil
}

func (r *VideoRepository) ListFilepathsByChannel(ctx context.Context, channelID int) ([]string, error) {
	filepaths := make([]string, 0)
	err := r.db.WithContext(ctx).
		Model(&models.Video{}).
		Where("channel_id = ?", channelID).
		Pluck("filepath", &filepaths).Error
	if err != nil {
		return nil, fmt.Errorf("list filepaths by channel: %w", err)
	}
	return filepaths, nil
}
