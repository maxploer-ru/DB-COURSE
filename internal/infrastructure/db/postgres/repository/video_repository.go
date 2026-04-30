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

func (r *VideoRepository) List(ctx context.Context, limit, offset int, sort domain.VideoSort) ([]*domain.Video, error) {
	var dbVideos []models.Video
	query := r.db.WithContext(ctx).Model(&models.Video{})
	query = applyVideoSort(query, sort)
	err := query.
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

func (r *VideoRepository) ListByChannel(ctx context.Context, channelID int, limit, offset int, sort domain.VideoSort) ([]*domain.Video, error) {
	var dbVideos []models.Video
	query := r.db.WithContext(ctx).Model(&models.Video{}).
		Where("channel_id = ?", channelID)
	query = applyVideoSort(query, sort)
	if err := query.
		Limit(limit).
		Offset(offset).
		Preload("Channel").
		Find(&dbVideos).Error; err != nil {
		return nil, fmt.Errorf("list videos by channel: %w", err)
	}

	domainVideos := make([]*domain.Video, 0, len(dbVideos)) // TODO: mapper ToDomainVideoList
	for _, dbVideo := range dbVideos {
		domainVideos = append(domainVideos, mappers.ToDomainVideo(&dbVideo))
	}
	return domainVideos, nil
}

func (r *VideoRepository) ListFilepathsByChannel(ctx context.Context, channelID int) ([]string, error) {
	filepath := make([]string, 0)
	err := r.db.WithContext(ctx).
		Model(&models.Video{}).
		Where("channel_id = ?", channelID).
		Pluck("filepath", &filepath).Error
	if err != nil {
		return nil, fmt.Errorf("list file paths by channel: %w", err)
	}
	return filepath, nil
}

func applyVideoSort(query *gorm.DB, sort domain.VideoSort) *gorm.DB {
	switch sort {
	case domain.VideoSortViews:
		return query.
			Joins("LEFT JOIN (SELECT video_id, COUNT(*) AS views_count FROM viewings GROUP BY video_id) v ON v.video_id = videos.id").
			Order("COALESCE(v.views_count, 0) DESC").
			Order("videos.created_at DESC")
	case domain.VideoSortRating:
		return query.
			Joins("LEFT JOIN (SELECT video_id, SUM(CASE WHEN liked THEN 1 ELSE -1 END) AS rating_score FROM video_ratings GROUP BY video_id) r ON r.video_id = videos.id").
			Order("COALESCE(r.rating_score, 0) DESC").
			Order("videos.created_at DESC")
	default:
		return query.Order("videos.created_at DESC")
	}
}
