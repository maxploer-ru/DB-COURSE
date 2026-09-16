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
	GetUserSubscriptions(ctx context.Context, userID int, limit, offset int) (*domain.PageResponse[*domain.Subscription], error)
	ResetNewVideosCount(ctx context.Context, userID, channelID int) error
	NotifyAboutNewVideo(ctx context.Context, channelID int) error
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

	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("get channel failed: %w", err)
	}
	if channel == nil {
		return domain.ErrChannelNotFound
	}
	if channel.UserID == userID {
		return domain.ErrSelfSubscription
	}

	created, err := s.subRepo.Subscribe(ctx, userID, channelID)
	if err != nil {
		return fmt.Errorf("subscribe to channel failed: %w", err)
	}

	if created {
		if err := s.counter.Increment(ctx, channelID); err != nil {
			logger.WarnContext(ctx, "Failed to increment subscriber cache, consistency may be delayed", slog.String("error", err.Error()))
		}
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

	deleted, err := s.subRepo.Unsubscribe(ctx, userID, channelID)
	if err != nil {
		return fmt.Errorf("unsubscribe to channel failed: %w", err)
	}

	if deleted {
		if err := s.counter.Decrement(ctx, channelID); err != nil {
			logger.WarnContext(ctx, "Failed to decrement subscriber cache, consistency may be delayed", slog.String("error", err.Error()))
		}
	}

	logger.InfoContext(ctx, "Subscription removed successfully")
	return nil
}

func (s *subscriptionService) NotifyAboutNewVideo(ctx context.Context, channelID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "SubscriptionService"),
		slog.String("operation", "NotifyAboutNewVideo"),
		slog.Int("channel_id", channelID),
	)

	if err := s.subRepo.NotifySubscribersAboutNewVideo(ctx, channelID); err != nil {
		logger.ErrorContext(ctx, "Failed to notify subscribers about new video", slog.String("error", err.Error()))
		return fmt.Errorf("notify subscribers failed: %w", err)
	}

	logger.InfoContext(ctx, "Subscribers notified about new video")
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
	cnt, hit, cacheErr := s.counter.Get(ctx, channelID)
	if cacheErr == nil && hit {
		logger.DebugContext(ctx, "Subscriber count retrieved from cache", slog.Int("count", cnt))
		return cnt, nil
	}
	if cacheErr != nil {
		logger.WarnContext(ctx, "Cache error, falling back to DB", slog.String("error", cacheErr.Error()))
	}

	logger.DebugContext(ctx, "Fetching subscriber count from database")
	realCnt, err := s.subRepo.GetSubscribersCount(ctx, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get subscriber count from DB", slog.String("error", err.Error()))
		return 0, err
	}

	if cacheErr == nil {
		_ = s.counter.Set(ctx, channelID, realCnt)
	}

	logger.DebugContext(ctx, "Subscriber count retrieved from DB and cache populated", slog.Int("count", realCnt))
	return realCnt, nil
}

func (s *subscriptionService) GetUserSubscriptions(ctx context.Context, userID int, limit, offset int) (*domain.PageResponse[*domain.Subscription], error) {
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
	total, err := s.subRepo.CountUserSubscriptions(ctx, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to count user subscriptions", slog.String("error", err.Error()))
		return nil, err
	}
	logger.DebugContext(ctx, "User subscriptions retrieved", slog.Int("count", len(subs)))
	return &domain.PageResponse[*domain.Subscription]{
		Items:      subs,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}, nil
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
