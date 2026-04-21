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

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	model := &models.Comment{
		UserID:    comment.UserID,
		VideoID:   comment.VideoID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	comment.ID = model.ID
	return nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id int) (*domain.Comment, error) {
	var model models.Comment
	err := r.db.WithContext(ctx).First(&model, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get comment by id: %w", err)
	}
	return mappers.ToDomainComment(&model), nil
}

func (r *CommentRepository) ListByVideo(ctx context.Context, videoID int, limit, offset int) ([]*domain.Comment, error) {
	var models []models.Comment
	err := r.db.WithContext(ctx).
		Where("video_id = ?", videoID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("list comments by video: %w", err)
	}
	return mappers.ToDomainCommentList(models), nil
}

func (r *CommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	model := mappers.FromDomainComment(comment)
	result := r.db.WithContext(ctx).Model(&models.Comment{}).
		Where("id = ?", model.ID).
		Update("content", model.Content)
	if result.Error != nil {
		return fmt.Errorf("update comment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *CommentRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&models.Comment{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete comment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *CommentRepository) CountByVideo(ctx context.Context, videoID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Comment{}).
		Where("video_id = ?", videoID).
		Count(&count).Error
	return count, err
}
