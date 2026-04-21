package repository

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/mappers"
	"ZVideo/internal/infrastructure/db/postgres/models"
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type VideoRatingRepository struct {
	db *gorm.DB
}

func NewVideoRatingRepository(db *gorm.DB) *VideoRatingRepository {
	return &VideoRatingRepository{db: db}
}

func (r *VideoRatingRepository) Create(ctx context.Context, rating *domain.VideoRating) error {
	model := &models.VideoRating{
		UserID:  rating.UserID,
		VideoID: rating.VideoID,
		Liked:   rating.Liked,
		RatedAt: time.Now(),
	}
	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrAlreadyRated
		}
		return fmt.Errorf("create video rating: %w", err)
	}
	return nil
}

func (r *VideoRatingRepository) Update(ctx context.Context, rating *domain.VideoRating) error {
	model := mappers.FromDomainVideoRating(rating)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *VideoRatingRepository) Delete(ctx context.Context, userID, videoID int) error {
	res := r.db.WithContext(ctx).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		Delete(&models.VideoRating{})
	if res.Error != nil {
		return fmt.Errorf("delete video rating: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrRatingNotFound
	}
	return nil
}

func (r *VideoRatingRepository) GetByUserAndVideo(ctx context.Context, userID, videoID int) (*domain.VideoRating, error) {
	var model models.VideoRating
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mappers.ToDomainVideoRating(&model), nil
}

func (r *VideoRatingRepository) GetStats(ctx context.Context, videoID int) (likes, dislikes int, err error) {
	var result struct {
		Likes    int
		Dislikes int
	}
	err = r.db.WithContext(ctx).
		Model(&models.VideoRating{}).
		Select("SUM(CASE WHEN liked THEN 1 ELSE 0 END) as likes, SUM(CASE WHEN NOT liked THEN 1 ELSE 0 END) as dislikes").
		Where("video_id = ?", videoID).
		Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}
	return result.Likes, result.Dislikes, nil
}
