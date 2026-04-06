package repositories

import (
	"ZVideo/internal/domain/channel/entity"
	"context"

	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{
		db: db,
	}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *entity.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *SubscriptionRepository) Delete(ctx context.Context, userID, channelID int) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND channel_id = ?", userID, channelID).
		Delete(&entity.Subscription{}).Error
}

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID int) ([]*entity.Subscription, error) {
	var subs []*entity.Subscription
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&subs).Error
	return subs, err
}

func (r *SubscriptionRepository) GetByChannelID(ctx context.Context, channelID int) ([]*entity.Subscription, error) {
	var subs []*entity.Subscription
	err := r.db.WithContext(ctx).Where("channel_id = ?", channelID).Find(&subs).Error
	return subs, err
}

func (r *SubscriptionRepository) IsSubscribed(ctx context.Context, userID, channelID int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Subscription{}).
		Where("user_id = ? AND channel_id = ?", userID, channelID).
		Count(&count).Error
	return count > 0, err
}

func (r *SubscriptionRepository) CountSubscribers(ctx context.Context, channelID int) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Subscription{}).
		Where("channel_id = ?", channelID).
		Count(&count).Error
	return int(count), err
}
