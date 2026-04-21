package repository

import (
	"ZVideo/internal/domain"
	"context"
)

type SubscriptionRepository interface {
	Subscribe(ctx context.Context, userID, channelID int) (created bool, err error)
	Unsubscribe(ctx context.Context, userID, channelID int) (deleted bool, err error)
	IsSubscribed(ctx context.Context, userID, channelID int) (bool, error)
	GetSubscribersCount(ctx context.Context, channelID int) (int, error)
	GetUserSubscriptions(ctx context.Context, userID int, limit, offset int) ([]*domain.Subscription, error)
	NotifySubscribersAboutNewVideo(ctx context.Context, channelID int) error
	ResetNewVideosCount(ctx context.Context, userID, channelID int) error
}
