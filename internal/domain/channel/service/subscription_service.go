package service

import (
	"ZVideo/internal/domain/channel/entity"
	"ZVideo/internal/domain/channel/repository"
	"context"
	"errors"
	"fmt"
)

var (
	ErrAlreadySubscribed = errors.New("already subscribed")
	ErrNotSubscribed     = errors.New("not subscribed")
	ErrChannelNotFound   = errors.New("channel not found")
)

type SubscriptionService interface {
	Subscribe(ctx context.Context, userID, channelID int) error
	Unsubscribe(ctx context.Context, userID, channelID int) error
	IsSubscribed(ctx context.Context, userID, channelID int) (bool, error)
	GetUserSubscriptions(ctx context.Context, userID int) ([]*entity.Subscription, error)
	GetChannelSubscribers(ctx context.Context, channelID int) ([]*entity.Subscription, error)
	GetSubscriberCount(ctx context.Context, channelID int) (int, error)
}

type subscriptionService struct {
	subRepo     repository.SubscriptionRepository
	channelRepo repository.ChannelRepository
}

func NewSubscriptionService(subRepo repository.SubscriptionRepository, channelRepo repository.ChannelRepository) SubscriptionService {
	return &subscriptionService{
		subRepo:     subRepo,
		channelRepo: channelRepo,
	}
}

func (s *subscriptionService) Subscribe(ctx context.Context, userID, channelID int) error {

	ch, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("get channel: %w", err)
	}
	if ch == nil {
		return ErrChannelNotFound
	}

	exists, err := s.subRepo.IsSubscribed(ctx, userID, channelID)
	if err != nil {
		return fmt.Errorf("check subscription: %w", err)
	}
	if exists {
		return ErrAlreadySubscribed
	}
	sub := &entity.Subscription{
		UserID:    userID,
		ChannelID: channelID,
	}
	if err := s.subRepo.Create(ctx, sub); err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}
	return nil
}

func (s *subscriptionService) Unsubscribe(ctx context.Context, userID, channelID int) error {
	exists, err := s.subRepo.IsSubscribed(ctx, userID, channelID)
	if err != nil {
		return fmt.Errorf("check subscription: %w", err)
	}
	if !exists {
		return ErrNotSubscribed
	}
	if err := s.subRepo.Delete(ctx, userID, channelID); err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	return nil
}

func (s *subscriptionService) IsSubscribed(ctx context.Context, userID, channelID int) (bool, error) {
	return s.subRepo.IsSubscribed(ctx, userID, channelID)
}

func (s *subscriptionService) GetUserSubscriptions(ctx context.Context, userID int) ([]*entity.Subscription, error) {
	return s.subRepo.GetByUserID(ctx, userID)
}

func (s *subscriptionService) GetChannelSubscribers(ctx context.Context, channelID int) ([]*entity.Subscription, error) {
	return s.subRepo.GetByChannelID(ctx, channelID)
}

func (s *subscriptionService) GetSubscriberCount(ctx context.Context, channelID int) (int, error) {
	return s.subRepo.CountSubscribers(ctx, channelID)
}
