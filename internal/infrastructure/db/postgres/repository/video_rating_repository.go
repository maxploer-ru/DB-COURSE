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
	"gorm.io/gorm/clause"
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
		if isUniqueViolation(err) {
			return domain.ErrAlreadyRated
		}
		return fmt.Errorf("create video rating: %w", err)
	}
	return nil
}

func (r *VideoRatingRepository) Update(ctx context.Context, rating *domain.VideoRating) error {
	result := r.db.WithContext(ctx).Model(&models.VideoRating{}).
		Where("user_id = ? AND video_id = ?", rating.UserID, rating.VideoID).
		Updates(map[string]any{"liked": rating.Liked, "rated_at": time.Now()})
	if result.Error != nil {
		return fmt.Errorf("update video rating: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrRatingNotFound
	}
	return nil
}

// Upsert atomically creates or changes a rating and returns its previous state.
// The advisory lock closes the missing-row race that a normal SELECT FOR UPDATE
// cannot protect against in PostgreSQL.
func (r *VideoRatingRepository) Upsert(ctx context.Context, rating *domain.VideoRating) (previous *domain.VideoRating, created bool, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?, ?)", rating.UserID, rating.VideoID).Error; err != nil {
			return err
		}
		var model models.VideoRating
		findErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND video_id = ?", rating.UserID, rating.VideoID).
			First(&model).Error
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			model = models.VideoRating{UserID: rating.UserID, VideoID: rating.VideoID, Liked: rating.Liked, RatedAt: time.Now()}
			if err := tx.Create(&model).Error; err != nil {
				return fmt.Errorf("create video rating: %w", err)
			}
			rating.RatedAt = model.RatedAt
			created = true
			return nil
		}
		if findErr != nil {
			return findErr
		}
		previous = mappers.ToDomainVideoRating(&model)
		model.Liked = rating.Liked
		model.RatedAt = time.Now()
		if err := tx.Model(&models.VideoRating{}).
			Where("user_id = ? AND video_id = ?", rating.UserID, rating.VideoID).
			Updates(map[string]any{"liked": model.Liked, "rated_at": model.RatedAt}).Error; err != nil {
			return err
		}
		rating.RatedAt = model.RatedAt
		return nil
	})
	return previous, created, err
}

func (r *VideoRatingRepository) DeleteAndGet(ctx context.Context, userID, videoID int) (previous *domain.VideoRating, found bool, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?, ?)", userID, videoID).Error; err != nil {
			return err
		}
		var model models.VideoRating
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND video_id = ?", userID, videoID).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		previous = mappers.ToDomainVideoRating(&model)
		found = true
		return tx.Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&models.VideoRating{}).Error
	})
	return previous, found, err
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
		Select("COALESCE(SUM(CASE WHEN liked THEN 1 ELSE 0 END), 0) as likes, COALESCE(SUM(CASE WHEN NOT liked THEN 1 ELSE 0 END), 0) as dislikes").
		Where("video_id = ?", videoID).
		Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}
	return result.Likes, result.Dislikes, nil
}
