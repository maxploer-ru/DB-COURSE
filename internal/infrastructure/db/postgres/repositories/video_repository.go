package repositories

import (
	"ZVideo/internal/domain/video/entity"
	"context"
	"errors"

	"gorm.io/gorm"
)

type VideoRepository struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) *VideoRepository {
	return &VideoRepository{
		db: db,
	}
}

func (r *VideoRepository) Create(ctx context.Context, video *entity.Video) error {
	return r.db.WithContext(ctx).Create(video).Error
}

func (r *VideoRepository) GetByID(ctx context.Context, id int) (*entity.Video, error) {
	var video entity.Video
	err := r.db.WithContext(ctx).First(&video, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &video, nil
}

func (r *VideoRepository) Update(ctx context.Context, video *entity.Video) error {
	return r.db.WithContext(ctx).Save(video).Error
}

func (r *VideoRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&entity.Video{}, id).Error
}

func (r *VideoRepository) GetByChannelID(ctx context.Context, channelID int, limit, offset int) ([]*entity.Video, error) {
	var videos []*entity.Video
	err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&videos).Error
	return videos, err
}

func (r *VideoRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entity.Video, error) {
	var videos []*entity.Video
	searchTerm := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Where("title ILIKE ? OR description ILIKE ?", searchTerm, searchTerm).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&videos).Error
	return videos, err
}
