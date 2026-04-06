package repositories

import (
	"ZVideo/internal/domain/comment/entity"
	"context"
	"errors"

	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{
		db: db,
	}
}

func (r *CommentRepository) Create(ctx context.Context, comment *entity.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *CommentRepository) GetByID(ctx context.Context, id int) (*entity.Comment, error) {
	var comment entity.Comment
	err := r.db.WithContext(ctx).First(&comment, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) GetByVideoID(ctx context.Context, videoID int, limit, offset int) ([]*entity.Comment, error) {
	var comments []*entity.Comment
	err := r.db.WithContext(ctx).
		Where("video_id = ?", videoID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&comments).Error
	return comments, err
}

func (r *CommentRepository) Update(ctx context.Context, comment *entity.Comment) error {
	return r.db.WithContext(ctx).Save(comment).Error
}

func (r *CommentRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&entity.Comment{}, id).Error
}

func (r *CommentRepository) GetCountByVideo(ctx context.Context, videoID int) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Comment{}).Where("video_id = ?", videoID).Count(&count).Error
	return int(count), err
}
