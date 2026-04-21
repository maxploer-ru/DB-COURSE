package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"fmt"
	"log/slog"
)

type SubscriptionService interface {
	Subscribe(ctx context.Context, userID, channelID int) error
	Unsubscribe(ctx context.Context, userID, channelID int) error
	IsSubscribed(ctx context.Context, userID, channelID int) (bool, error)
	GetSubscribersCount(ctx context.Context, channelID int) (int, error)
	GetUserSubscriptions(ctx context.Context, userID int, limit, offset int) ([]*domain.Subscription, error)
	ResetNewVideosCount(ctx context.Context, userID, channelID int) error
}

type subscriptionService struct {
	subRepo     repository.SubscriptionRepository
	channelRepo repository.ChannelRepository
	counter     repository.SubscriberCounter
}

func NewSubscriptionService(
	subRepo repository.SubscriptionRepository,
	channelRepo repository.ChannelRepository,
	counter repository.SubscriberCounter) SubscriptionService {
	return &subscriptionService{
		subRepo:     subRepo,
		channelRepo: channelRepo,
		counter:     counter,
	}
}

func (s *subscriptionService) Subscribe(ctx context.Context, userID, channelID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "SubscriptionService"),
		slog.String("operation", "Subscribe"),
		slog.Int("user_id", userID),
		slog.Int("channel_id", channelID),
	)

	logger.DebugContext(ctx, "Checking channel existence")
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get channel", slog.String("error", err.Error()))
		return fmt.Errorf("get channel failed: %w", err)
	}
	if channel == nil {
		logger.WarnContext(ctx, "Channel not found")
		return domain.ErrChannelNotFound
	}
	if channel.UserID == userID {
		logger.WarnContext(ctx, "Attempt to self-subscribe")
		return domain.ErrSelfSubscription
	}

	logger.DebugContext(ctx, "Creating subscription in repository")
	created, err := s.subRepo.Subscribe(ctx, userID, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create subscription", slog.String("error", err.Error()))
		return fmt.Errorf("subscribe to channel failed: %w", err)
	}
	if created {
		_ = s.counter.Increment(ctx, channelID)
	} else {
		logger.DebugContext(ctx, "Subscription already exists, skip counter increment")
	}
	logger.InfoContext(ctx, "Subscription created successfully")
	return nil
}

func (s *subscriptionService) Unsubscribe(ctx context.Context, userID, channelID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "SubscriptionService"),
		slog.String("operation", "Unsubscribe"),
		slog.Int("user_id", userID),
		slog.Int("channel_id", channelID),
	)

	logger.DebugContext(ctx, "Deleting subscription from repository")
	deleted, err := s.subRepo.Unsubscribe(ctx, userID, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete subscription", slog.String("error", err.Error()))
		return fmt.Errorf("unsubscribe to channel failed: %w", err)
	}
	if deleted {
		_ = s.counter.Decrement(ctx, channelID)
	} else {
		logger.DebugContext(ctx, "Subscription does not exist, skip counter decrement")
	}
	logger.InfoContext(ctx, "Subscription removed successfully")
	return nil
}

func (s *subscriptionService) IsSubscribed(ctx context.Context, userID, channelID int) (bool, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "SubscriptionService"),
		slog.String("operation", "IsSubscribed"),
		slog.Int("user_id", userID),
		slog.Int("channel_id", channelID),
	)

	logger.DebugContext(ctx, "Checking subscription status")
	subscribed, err := s.subRepo.IsSubscribed(ctx, userID, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check subscription status", slog.String("error", err.Error()))
		return false, err
	}
	logger.DebugContext(ctx, "Subscription status checked", slog.Bool("subscribed", subscribed))
	return subscribed, nil
}

func (s *subscriptionService) GetSubscribersCount(ctx context.Context, channelID int) (int, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "SubscriptionService"),
		slog.String("operation", "GetSubscribersCount"),
		slog.Int("channel_id", channelID),
	)

	logger.DebugContext(ctx, "Trying to get subscriber count from cache")
	cnt, hit, err := s.counter.Get(ctx, channelID)
	if err == nil && hit {
		logger.DebugContext(ctx, "Subscriber count retrieved from cache", slog.Int("count", cnt))
		return cnt, nil
	}
	if err != nil {
		logger.WarnContext(ctx, "Cache error, falling back to DB", slog.String("error", err.Error()))
	}

	logger.DebugContext(ctx, "Fetching subscriber count from database")
	realCnt, err := s.subRepo.GetSubscribersCount(ctx, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get subscriber count from DB", slog.String("error", err.Error()))
		return 0, err
	}

	_ = s.counter.Set(ctx, channelID, realCnt)
	logger.DebugContext(ctx, "Subscriber count retrieved from DB and cache populated", slog.Int("count", realCnt))
	return realCnt, nil
}

func (s *subscriptionService) GetUserSubscriptions(ctx context.Context, userID int, limit, offset int) ([]*domain.Subscription, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "SubscriptionService"),
		slog.String("operation", "GetUserSubscriptions"),
		slog.Int("user_id", userID),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	logger.DebugContext(ctx, "Fetching user subscriptions from repository")
	subs, err := s.subRepo.GetUserSubscriptions(ctx, userID, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get user subscriptions", slog.String("error", err.Error()))
		return nil, err
	}
	logger.DebugContext(ctx, "User subscriptions retrieved", slog.Int("count", len(subs)))
	return subs, nil
}

func (s *subscriptionService) ResetNewVideosCount(ctx context.Context, userID, channelID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "SubscriptionService"),
		slog.String("operation", "ResetNewVideosCount"),
		slog.Int("user_id", userID),
		slog.Int("channel_id", channelID),
	)

	if err := s.subRepo.ResetNewVideosCount(ctx, userID, channelID); err != nil {
		logger.ErrorContext(ctx, "Failed to reset new videos count", slog.String("error", err.Error()))
		return err
	}
	logger.DebugContext(ctx, "New videos count reset")
	return nil
}
