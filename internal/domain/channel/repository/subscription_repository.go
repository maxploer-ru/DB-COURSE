package repository

import (
	"ZVideo/internal/domain/channel/entity"
	"context"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *entity.Subscription) error
	Delete(ctx context.Context, userID, channelID int) error
	GetByUserID(ctx context.Context, userID int) ([]*entity.Subscription, error)
	GetByChannelID(ctx context.Context, channelID int) ([]*entity.Subscription, error)
	IsSubscribed(ctx context.Context, userID, channelID int) (bool, error)
	CountSubscribers(ctx context.Context, channelID int) (int, error)
}
