package repository

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/mappers"
	"ZVideo/internal/infrastructure/db/postgres/models"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Subscribe(ctx context.Context, userID, channelID int) (bool, error) {
	sub := &models.Subscription{
		UserID:         userID,
		ChannelID:      channelID,
		NewVideosCount: 0,
		SubscribedAt:   time.Now(),
	}
	err := r.db.WithContext(ctx).Create(sub).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "duplicate key") {
			return false, nil
		}
		return false, fmt.Errorf("subscribe failed: %w", err)
	}
	return true, nil
}

func (r *SubscriptionRepository) Unsubscribe(ctx context.Context, userID, channelID int) (bool, error) {
	res := r.db.WithContext(ctx).
		Where("user_id = ? AND channel_id = ?", userID, channelID).
		Delete(&models.Subscription{})
	if res.Error != nil {
		return false, fmt.Errorf("unsubscribe failed: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func (r *SubscriptionRepository) IsSubscribed(ctx context.Context, userID, channelID int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("user_id = ? AND channel_id = ?", userID, channelID).
		Count(&count).Error
	return count > 0, err
}

func (r *SubscriptionRepository) GetSubscribersCount(ctx context.Context, channelID int) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("channel_id = ?", channelID).
		Count(&count).Error
	return int(count), err
}

func (r *SubscriptionRepository) GetUserSubscriptions(ctx context.Context, userID int, limit, offset int) ([]*domain.Subscription, error) {
	var subs []models.Subscription
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("subscribed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&subs).Error
	if err != nil {
		return nil, err
	}
	return mappers.ToDomainSubscriptionList(subs), nil
}

func (r *SubscriptionRepository) NotifySubscribersAboutNewVideo(ctx context.Context, channelID int) error {
	if err := r.db.WithContext(ctx).
		Exec("SELECT notify_subscribers_about_new_video(?)", channelID).Error; err != nil {
		return fmt.Errorf("notify subscribers failed: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) ResetNewVideosCount(ctx context.Context, userID, channelID int) error {
	res := r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("user_id = ? AND channel_id = ?", userID, channelID).
		Update("new_videos_count", 0)
	if res.Error != nil {
		return fmt.Errorf("reset new videos count failed: %w", res.Error)
	}
	return nil
}
