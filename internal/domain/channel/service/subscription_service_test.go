package service_test

import (
	"ZVideo/internal/domain/channel/entity"
	"ZVideo/internal/domain/channel/repository/mocks"
	"ZVideo/internal/domain/channel/service"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSubscriptionService_Subscribe_Success(t *testing.T) {
	mockSubRepo := new(mocks.MockSubscriptionRepository)
	mockChanRepo := new(mocks.MockChannelRepository)
	svc := service.NewSubscriptionService(mockSubRepo, mockChanRepo)

	channel := &entity.Channel{ID: 1}
	mockChanRepo.On("GetByID", mock.Anything, 1).Return(channel, nil)
	mockSubRepo.On("IsSubscribed", mock.Anything, 1, 1).Return(false, nil)
	mockSubRepo.On("Create", mock.Anything, mock.MatchedBy(func(sub *entity.Subscription) bool {
		return sub.UserID == 1 && sub.ChannelID == 1
	})).Return(nil)

	err := svc.Subscribe(context.Background(), 1, 1)
	assert.NoError(t, err)
	mockChanRepo.AssertExpectations(t)
	mockSubRepo.AssertExpectations(t)
}

func TestSubscriptionService_Subscribe_AlreadySubscribed(t *testing.T) {
	mockSubRepo := new(mocks.MockSubscriptionRepository)
	mockChanRepo := new(mocks.MockChannelRepository)
	svc := service.NewSubscriptionService(mockSubRepo, mockChanRepo)

	channel := &entity.Channel{ID: 1}
	mockChanRepo.On("GetByID", mock.Anything, 1).Return(channel, nil)
	mockSubRepo.On("IsSubscribed", mock.Anything, 1, 1).Return(true, nil)

	err := svc.Subscribe(context.Background(), 1, 1)
	assert.Error(t, err)
	assert.Equal(t, service.ErrAlreadySubscribed, err)
	mockSubRepo.AssertNotCalled(t, "Create")
}

func TestSubscriptionService_Unsubscribe_Success(t *testing.T) {
	mockSubRepo := new(mocks.MockSubscriptionRepository)
	mockChanRepo := new(mocks.MockChannelRepository)
	svc := service.NewSubscriptionService(mockSubRepo, mockChanRepo)

	mockSubRepo.On("IsSubscribed", mock.Anything, 1, 1).Return(true, nil)
	mockSubRepo.On("Delete", mock.Anything, 1, 1).Return(nil)

	err := svc.Unsubscribe(context.Background(), 1, 1)
	assert.NoError(t, err)
	mockSubRepo.AssertExpectations(t)
}

func TestSubscriptionService_GetSubscriberCount(t *testing.T) {
	mockSubRepo := new(mocks.MockSubscriptionRepository)
	mockChanRepo := new(mocks.MockChannelRepository)
	svc := service.NewSubscriptionService(mockSubRepo, mockChanRepo)

	mockSubRepo.On("CountSubscribers", mock.Anything, 1).Return(5, nil)

	cnt, err := svc.GetSubscriberCount(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, 5, cnt)
}
