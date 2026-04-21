package repository

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/mappers"
	"ZVideo/internal/infrastructure/db/postgres/models"
	"context"
	"errors"

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
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *ChannelRepository) GetByID(ctx context.Context, id int) (*domain.Channel, error) {
	var model models.Channel
	err := r.db.WithContext(ctx).First(&model, id).Error
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
	var channel models.Channel
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&channel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mappers.ToDomainChannel(&channel), nil
}

func (r *ChannelRepository) Update(ctx context.Context, channel *domain.Channel) error {
	return r.db.WithContext(ctx).Save(mappers.FromDomainChannel(channel)).Error
}

func (r *ChannelRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Channel{}, id).Error
}

func (r *ChannelRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Channel{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}
