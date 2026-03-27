package mocks

import (
	"ZVideo/internal/domain/channel/entity"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, sub *entity.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) Delete(ctx context.Context, userID, channelID int) error {
	args := m.Called(ctx, userID, channelID)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetByUserID(ctx context.Context, userID int) ([]*entity.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) GetByChannelID(ctx context.Context, channelID int) ([]*entity.Subscription, error) {
	args := m.Called(ctx, channelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) IsSubscribed(ctx context.Context, userID, channelID int) (bool, error) {
	args := m.Called(ctx, userID, channelID)
	return args.Bool(0), args.Error(1)
}

func (m *MockSubscriptionRepository) CountSubscribers(ctx context.Context, channelID int) (int, error) {
	args := m.Called(ctx, channelID)
	return args.Int(0), args.Error(1)
}
