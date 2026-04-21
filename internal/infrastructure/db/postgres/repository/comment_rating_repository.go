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
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrAlreadyRated
		}
		return fmt.Errorf("create comment rating: %w", err)
	}
	return nil
}

func (r *CommentRatingRepository) Update(ctx context.Context, rating *domain.CommentRating) error {
	model := mappers.FromDomainCommentRating(rating)
	return r.db.WithContext(ctx).Save(model).Error
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
		Select("SUM(CASE WHEN liked THEN 1 ELSE 0 END) as likes, SUM(CASE WHEN NOT liked THEN 1 ELSE 0 END) as dislikes").
		Where("comment_id = ?", commentID).
		Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}
	return result.Likes, result.Dislikes, nil
}
