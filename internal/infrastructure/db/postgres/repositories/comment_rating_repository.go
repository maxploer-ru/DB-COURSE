package repositories

import (
	"ZVideo/internal/domain/comment/entity"
	"context"
	"errors"

	"gorm.io/gorm"
)

type CommentRatingRepository struct {
	db *gorm.DB
}

func NewCommentRatingRepository(db *gorm.DB) *CommentRatingRepository {
	return &CommentRatingRepository{
		db: db,
	}
}

func (r *CommentRatingRepository) Create(ctx context.Context, rating *entity.CommentRating) error {
	return r.db.WithContext(ctx).Create(rating).Error
}

func (r *CommentRatingRepository) Update(ctx context.Context, rating *entity.CommentRating) error {
	return r.db.WithContext(ctx).Save(rating).Error
}

func (r *CommentRatingRepository) Delete(ctx context.Context, userID, commentID int) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND comment_id = ?", userID, commentID).
		Delete(&entity.CommentRating{}).Error
}

func (r *CommentRatingRepository) GetByUserAndComment(ctx context.Context, userID, commentID int) (*entity.CommentRating, error) {
	var rating entity.CommentRating
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND comment_id = ?", userID, commentID).
		First(&rating).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rating, nil
}

func (r *CommentRatingRepository) GetCommentRatingStats(ctx context.Context, commentID int) (likes, dislikes int, err error) {
	type result struct {
		IsLike bool
		Count  int
	}
	var results []result

	err = r.db.WithContext(ctx).Model(&entity.CommentRating{}).
		Select("is_like, COUNT(*) as count").
		Where("comment_id = ?", commentID).
		Group("is_like").
		Scan(&results).Error

	if err != nil {
		return 0, 0, err
	}

	for _, res := range results {
		if res.IsLike {
			likes = res.Count
		} else {
			dislikes = res.Count
		}
	}

	return likes, dislikes, nil
}
