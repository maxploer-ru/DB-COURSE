package repositories

import (
	"ZVideo/internal/domain/channel/entity"
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

func (r *ChannelRepository) Create(ctx context.Context, channel *entity.Channel) error {
	return r.db.WithContext(ctx).Create(channel).Error
}

func (r *ChannelRepository) GetByID(ctx context.Context, id int) (*entity.Channel, error) {
	var channel entity.Channel
	err := r.db.WithContext(ctx).First(&channel, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &channel, nil
}

func (r *ChannelRepository) GetByUserID(ctx context.Context, userID int) ([]*entity.Channel, error) {
	var channels []*entity.Channel
	err := r.db.WithContext(ctx).
		Where("owner_id = ?", userID).
		Order("created_at DESC").
		Find(&channels).Error
	return channels, err
}

func (r *ChannelRepository) GetByName(ctx context.Context, name string) (*entity.Channel, error) {
	var channel entity.Channel
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&channel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &channel, nil
}

func (r *ChannelRepository) Update(ctx context.Context, channel *entity.Channel) error {
	return r.db.WithContext(ctx).Save(channel).Error
}

func (r *ChannelRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&entity.Channel{}, id).Error
}

func (r *ChannelRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Channel{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}
