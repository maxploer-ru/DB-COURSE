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

type CommunityRepository struct {
	db *gorm.DB
}

func NewCommunityRepository(db *gorm.DB) *CommunityRepository {
	return &CommunityRepository{db: db}
}

func (r *CommunityRepository) CreatePost(ctx context.Context, post *domain.CommunityPost) error {
	model := mappers.FromDomainCommunityPost(post)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("create community post: %w", err)
	}
	post.ID = model.ID
	return nil
}

func (r *CommunityRepository) GetPostByID(ctx context.Context, id int) (*domain.CommunityPost, error) {
	var model models.CommunityPost
	err := r.db.WithContext(ctx).First(&model, id).
		Preload("User").Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get community post by id: %w", err)
	}
	return mappers.ToDomainCommunityPost(&model), nil
}

func (r *CommunityRepository) ListPostsByChannel(ctx context.Context, channelID int, limit, offset int) ([]*domain.CommunityPost, error) {
	var modelsList []models.CommunityPost
	err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&modelsList).Error
	if err != nil {
		return nil, fmt.Errorf("list community posts by channel: %w", err)
	}
	return mappers.ToDomainCommunityPostList(modelsList), nil
}

func (r *CommunityRepository) CountPostsByChannel(ctx context.Context, channelID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CommunityPost{}).
		Where("channel_id = ?", channelID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count community posts by channel: %w", err)
	}
	return count, nil
}

func (r *CommunityRepository) UpdatePost(ctx context.Context, post *domain.CommunityPost) error {
	model := mappers.FromDomainCommunityPost(post)
	result := r.db.WithContext(ctx).Model(&models.CommunityPost{}).
		Where("id = ?", model.ID).
		Update("content", model.Content)
	if result.Error != nil {
		return fmt.Errorf("update community post: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *CommunityRepository) DeletePost(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&models.CommunityPost{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete community post: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *CommunityRepository) CreateComment(ctx context.Context, comment *domain.CommunityComment) error {
	model := mappers.FromDomainCommunityComment(comment)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("create community comment: %w", err)
	}
	comment.ID = model.ID
	return nil
}

func (r *CommunityRepository) GetCommentByID(ctx context.Context, id int) (*domain.CommunityComment, error) {
	var model models.CommunityComment
	err := r.db.WithContext(ctx).First(&model, id).
		Preload("User").Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get community comment by id: %w", err)
	}
	return mappers.ToDomainCommunityComment(&model), nil
}

func (r *CommunityRepository) ListCommentsByPost(ctx context.Context, postID int, limit, offset int) ([]*domain.CommunityComment, error) {
	var modelsList []models.CommunityComment
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("post_id = ?", postID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&modelsList).Error
	if err != nil {
		return nil, fmt.Errorf("list community comments by post: %w", err)
	}
	return mappers.ToDomainCommunityCommentList(modelsList), nil
}

func (r *CommunityRepository) CountCommentsByPost(ctx context.Context, postID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CommunityComment{}).
		Where("post_id = ?", postID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count community comments by post: %w", err)
	}
	return count, nil
}

func (r *CommunityRepository) UpdateComment(ctx context.Context, comment *domain.CommunityComment) error {
	model := mappers.FromDomainCommunityComment(comment)
	result := r.db.WithContext(ctx).Model(&models.CommunityComment{}).
		Where("id = ?", model.ID).
		Update("content", model.Content)
	if result.Error != nil {
		return fmt.Errorf("update community comment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *CommunityRepository) DeleteComment(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&models.CommunityComment{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete community comment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
