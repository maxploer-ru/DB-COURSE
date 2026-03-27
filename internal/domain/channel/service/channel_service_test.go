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

func TestChannelService_CreateChannel_Success(t *testing.T) {
	mockRepo := new(mocks.MockChannelRepository)
	svc := service.NewChannelService(mockRepo)

	mockRepo.On("ExistsByName", mock.Anything, "MyChannel").Return(false, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(ch *entity.Channel) bool {
		return ch.UserID == 1 && ch.Name == "MyChannel"
	})).Return(nil)

	channel, err := svc.CreateChannel(context.Background(), 1, "MyChannel", "Cool channel")
	assert.NoError(t, err)
	assert.NotNil(t, channel)
	assert.Equal(t, "MyChannel", channel.Name)
	mockRepo.AssertExpectations(t)
}

func TestChannelService_CreateChannel_Exists(t *testing.T) {
	mockRepo := new(mocks.MockChannelRepository)
	svc := service.NewChannelService(mockRepo)

	mockRepo.On("ExistsByName", mock.Anything, "Existing").Return(true, nil)

	_, err := svc.CreateChannel(context.Background(), 1, "Existing", "desc")
	assert.Error(t, err)
	assert.Equal(t, service.ErrChannelNameExists, err)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestChannelService_UpdateChannel_Success(t *testing.T) {
	mockRepo := new(mocks.MockChannelRepository)
	svc := service.NewChannelService(mockRepo)

	existing := &entity.Channel{ID: 1, UserID: 1, Name: "OldName", Description: "OldDesc"}
	mockRepo.On("GetByID", mock.Anything, 1).Return(existing, nil)
	mockRepo.On("ExistsByName", mock.Anything, "NewName").Return(false, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(ch *entity.Channel) bool {
		return ch.Name == "NewName" && ch.Description == "NewDesc"
	})).Return(nil)

	updated, err := svc.UpdateChannel(context.Background(), 1, strPtr("NewName"), strPtr("NewDesc"))
	assert.NoError(t, err)
	assert.Equal(t, "NewName", updated.Name)
	assert.Equal(t, "NewDesc", updated.Description)
}

func TestChannelService_DeleteChannel_NotOwner(t *testing.T) {
	mockRepo := new(mocks.MockChannelRepository)
	svc := service.NewChannelService(mockRepo)

	existing := &entity.Channel{ID: 1, UserID: 2}
	mockRepo.On("GetByID", mock.Anything, 1).Return(existing, nil)

	err := svc.DeleteChannel(context.Background(), 1, 3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not authorized")
}

func strPtr(s string) *string { return &s }
