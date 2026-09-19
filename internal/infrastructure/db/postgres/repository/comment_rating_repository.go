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

type CommentRatingRepository struct {
	db *gorm.DB
}

func NewCommentRatingRepository(db *gorm.DB) *CommentRatingRepository {
	return &CommentRatingRepository{db: db}
}

func (r *CommentRatingRepository) Create(ctx context.Context, rating *domain.CommentRating) error {
	model := &models.CommentRating{
		UserID:    rating.UserID,
		CommentID: rating.CommentID,
		Liked:     rating.Liked,
		RatedAt:   time.Now(),
	}
	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyRated
		}
		return fmt.Errorf("create comment rating: %w", err)
	}
	return nil
}

func (r *CommentRatingRepository) Update(ctx context.Context, rating *domain.CommentRating) error {
	result := r.db.WithContext(ctx).Model(&models.CommentRating{}).
		Where("user_id = ? AND comment_id = ?", rating.UserID, rating.CommentID).
		Updates(map[string]any{"liked": rating.Liked, "rated_at": time.Now()})
	if result.Error != nil {
		return fmt.Errorf("update comment rating: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrCommentRatingNotFound
	}
	return nil
}

func (r *CommentRatingRepository) Upsert(ctx context.Context, rating *domain.CommentRating) (previous *domain.CommentRating, created bool, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?, ?)", rating.UserID, rating.CommentID).Error; err != nil {
			return err
		}
		var model models.CommentRating
		findErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND comment_id = ?", rating.UserID, rating.CommentID).
			First(&model).Error
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			model = models.CommentRating{UserID: rating.UserID, CommentID: rating.CommentID, Liked: rating.Liked, RatedAt: time.Now()}
			if err := tx.Create(&model).Error; err != nil {
				return fmt.Errorf("create comment rating: %w", err)
			}
			rating.RatedAt = model.RatedAt
			created = true
			return nil
		}
		if findErr != nil {
			return findErr
		}
		previous = mappers.ToDomainCommentRating(&model)
		model.Liked = rating.Liked
		model.RatedAt = time.Now()
		return tx.Model(&models.CommentRating{}).
			Where("user_id = ? AND comment_id = ?", rating.UserID, rating.CommentID).
			Updates(map[string]any{"liked": model.Liked, "rated_at": model.RatedAt}).Error
	})
	return previous, created, err
}

func (r *CommentRatingRepository) DeleteAndGet(ctx context.Context, userID, commentID int) (previous *domain.CommentRating, found bool, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?, ?)", userID, commentID).Error; err != nil {
			return err
		}
		var model models.CommentRating
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND comment_id = ?", userID, commentID).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		previous = mappers.ToDomainCommentRating(&model)
		found = true
		return tx.Where("user_id = ? AND comment_id = ?", userID, commentID).Delete(&models.CommentRating{}).Error
	})
	return previous, found, err
}

func (r *CommentRatingRepository) Delete(ctx context.Context, userID, commentID int) error {
	res := r.db.WithContext(ctx).
		Where("user_id = ? AND comment_id = ?", userID, commentID).
		Delete(&models.CommentRating{})
	if res.Error != nil {
		return fmt.Errorf("delete comment rating: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrCommentRatingNotFound
	}
	return nil
}

func (r *CommentRatingRepository) GetByUserAndComment(ctx context.Context, userID, commentID int) (*domain.CommentRating, error) {
	var model models.CommentRating
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND comment_id = ?", userID, commentID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mappers.ToDomainCommentRating(&model), nil
}

func (r *CommentRatingRepository) GetStats(ctx context.Context, commentID int) (likes, dislikes int64, err error) {
	var result struct {
		Likes    int64
		Dislikes int64
	}
	err = r.db.WithContext(ctx).
		Model(&models.CommentRating{}).
		Select("COALESCE(SUM(CASE WHEN liked THEN 1 ELSE 0 END), 0) as likes, COALESCE(SUM(CASE WHEN NOT liked THEN 1 ELSE 0 END), 0) as dislikes").
		Where("comment_id = ?", commentID).
		Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}
	return result.Likes, result.Dislikes, nil
}
