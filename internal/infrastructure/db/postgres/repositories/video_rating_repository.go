package repositories

import (
	"ZVideo/internal/domain/video/entity"
	"context"
	"errors"

	"gorm.io/gorm"
)

type VideoRatingRepository struct {
	db *gorm.DB
}

func NewVideoRatingRepository(db *gorm.DB) *VideoRatingRepository {
	return &VideoRatingRepository{
		db: db,
	}
}

func (r *VideoRatingRepository) Create(ctx context.Context, rating *entity.VideoRating) error {
	return r.db.WithContext(ctx).Create(rating).Error
}

func (r *VideoRatingRepository) Update(ctx context.Context, rating *entity.VideoRating) error {
	return r.db.WithContext(ctx).Save(rating).Error
}

func (r *VideoRatingRepository) Delete(ctx context.Context, userID, videoID int) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		Delete(&entity.VideoRating{}).Error
}

func (r *VideoRatingRepository) GetByUserAndVideo(ctx context.Context, userID, videoID int) (*entity.VideoRating, error) {
	var rating entity.VideoRating
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		First(&rating).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rating, nil
}
