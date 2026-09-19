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

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{
		db: db,
	}
}

func (r *ChannelRepository) Create(ctx context.Context, channel *domain.Channel) error {
	model := mappers.FromDomainChannel(channel)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if mapped := mapUniqueConstraint(err, map[string]error{
			"channels_user_id_key": domain.ErrChannelAlreadyExists,
			"channels_name_key":    domain.ErrChannelNameAlreadyExists,
		}); mapped != nil {
			return mapped
		}
		return fmt.Errorf("create channel: %w", err)
	}
	channel.ID = model.ID
	return nil
}

func (r *ChannelRepository) GetByID(ctx context.Context, id int) (*domain.Channel, error) {
	var model models.Channel
	err := r.db.WithContext(ctx).Preload("User").First(&model, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mappers.ToDomainChannel(&model), nil
}

func (r *ChannelRepository) GetByUserID(ctx context.Context, userID int) (*domain.Channel, error) {
	var model models.Channel
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ?", userID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mappers.ToDomainChannel(&model), nil
}

func (r *ChannelRepository) GetByName(ctx context.Context, name string) (*domain.Channel, error) {
	var model models.Channel
	err := r.db.WithContext(ctx).Preload("User").Where("name = ?", name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mappers.ToDomainChannel(&model), nil
}

func (r *ChannelRepository) Update(ctx context.Context, channel *domain.Channel) error {
	result := r.db.WithContext(ctx).Model(&models.Channel{}).
		Where("id = ? AND user_id = ?", channel.ID, channel.UserID).
		Updates(map[string]any{
			"name":        channel.Name,
			"description": channel.Description,
		})
	if result.Error != nil {
		if mapped := mapUniqueConstraint(result.Error, map[string]error{
			"channels_name_key": domain.ErrChannelNameAlreadyExists,
		}); mapped != nil {
			return mapped
		}
		return fmt.Errorf("update channel: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrChannelNotFound
	}
	return nil
}

func (r *ChannelRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Channel{}, id).Error
}

func (r *ChannelRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Channel{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *ChannelRepository) ListChannels(ctx context.Context, limit, offset int) ([]*domain.Channel, error) {
	var dbModels []*models.Channel
	err := r.db.WithContext(ctx).
		Preload("User").
		Order("created_at DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbModels).Error
	if err != nil {
		return nil, err
	}
	return mappers.ToDomainChannelList(dbModels), nil
}

func (r *ChannelRepository) CountChannels(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Channel{}).
		Count(&count).Error
	return count, err
}
