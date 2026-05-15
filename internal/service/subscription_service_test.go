package service_test

import (
	"ZVideo/internal/domain"
	service "ZVideo/internal/service"
	"ZVideo/mocks"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubscriptionService_Subscribe(t *testing.T) {
	ctx := context.Background()
	subRepo := mocks.NewSubscriptionRepository(t)
	channelRepo := mocks.NewChannelRepository(t)
	counter := mocks.NewSubscriberCounter(t)

	channelRepo.On("GetByID", ctx, 3).Return(&domain.Channel{ID: 3, UserID: 10}, nil)
	subRepo.On("Subscribe", ctx, 1, 3).Return(true, nil)
	counter.On("Increment", ctx, 3).Return(nil)

	svc := service.NewSubscriptionService(subRepo, channelRepo, counter)
	err := svc.Subscribe(ctx, 1, 3)
	require.NoError(t, err)
}

func TestSubscriptionService_Unsubscribe(t *testing.T) {
	ctx := context.Background()
	subRepo := mocks.NewSubscriptionRepository(t)
	channelRepo := mocks.NewChannelRepository(t)
	counter := mocks.NewSubscriberCounter(t)

	subRepo.On("Unsubscribe", ctx, 1, 3).Return(true, nil)
	counter.On("Decrement", ctx, 3).Return(nil)

	svc := service.NewSubscriptionService(subRepo, channelRepo, counter)
	err := svc.Unsubscribe(ctx, 1, 3)
	require.NoError(t, err)
}

func TestSubscriptionService_IsSubscribed(t *testing.T) {
	ctx := context.Background()
	subRepo := mocks.NewSubscriptionRepository(t)
	channelRepo := mocks.NewChannelRepository(t)
	counter := mocks.NewSubscriberCounter(t)

	subRepo.On("IsSubscribed", ctx, 1, 3).Return(true, nil)

	svc := service.NewSubscriptionService(subRepo, channelRepo, counter)
	ok, err := svc.IsSubscribed(ctx, 1, 3)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestSubscriptionService_GetSubscribersCount(t *testing.T) {
	ctx := context.Background()
	subRepo := mocks.NewSubscriptionRepository(t)
	channelRepo := mocks.NewChannelRepository(t)
	counter := mocks.NewSubscriberCounter(t)

	counter.On("Get", ctx, 3).Return(5, true, nil)

	svc := service.NewSubscriptionService(subRepo, channelRepo, counter)
	count, err := svc.GetSubscribersCount(ctx, 3)
	require.NoError(t, err)
	require.Equal(t, 5, count)
}

func TestSubscriptionService_GetUserSubscriptions(t *testing.T) {
	ctx := context.Background()
	subRepo := mocks.NewSubscriptionRepository(t)
	channelRepo := mocks.NewChannelRepository(t)
	counter := mocks.NewSubscriberCounter(t)

	subRepo.On("GetUserSubscriptions", ctx, 1, 10, 0).Return([]*domain.Subscription{{UserID: 1}}, nil)

	svc := service.NewSubscriptionService(subRepo, channelRepo, counter)
	subs, err := svc.GetUserSubscriptions(ctx, 1, 10, 0)
	require.NoError(t, err)
	require.Len(t, subs, 1)
}

func TestSubscriptionService_ResetNewVideosCount(t *testing.T) {
	ctx := context.Background()
	subRepo := mocks.NewSubscriptionRepository(t)
	channelRepo := mocks.NewChannelRepository(t)
	counter := mocks.NewSubscriberCounter(t)

	subRepo.On("ResetNewVideosCount", ctx, 1, 3).Return(nil)

	svc := service.NewSubscriptionService(subRepo, channelRepo, counter)
	err := svc.ResetNewVideosCount(ctx, 1, 3)
	require.NoError(t, err)
}
