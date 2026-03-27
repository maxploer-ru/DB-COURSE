package service_test

import (
	"ZVideo/internal/domain/video/repository/mocks"
	"ZVideo/internal/domain/video/service"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type ChannelCheckerMock struct {
	mock.Mock
}

func (m *ChannelCheckerMock) Exists(ctx context.Context, channelID int) (bool, error) {
	args := m.Called(ctx, channelID)
	return args.Bool(0), args.Error(1)
}

func (m *ChannelCheckerMock) IsOwner(ctx context.Context, channelID, userID int) (bool, error) {
	args := m.Called(ctx, channelID, userID)
	return args.Bool(0), args.Error(1)
}

func TestCreateVideo_Success(t *testing.T) {
	videoRepo := new(mocks.VideoRepositoryMock)
	viewingRepo := new(mocks.ViewingRepositoryMock)
	cache := new(mocks.VideoStatsCacheMock)
	channel := new(ChannelCheckerMock)

	svc := service.NewVideoService(videoRepo, viewingRepo, cache, channel)

	channel.On("Exists", mock.Anything, 1).Return(true, nil)
	channel.On("IsOwner", mock.Anything, 1, 1).Return(true, nil)

	videoRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	video, err := svc.CreateVideo(context.Background(), 1, 1, "title", "desc", "path")

	assert.NoError(t, err)
	assert.NotNil(t, video)
}
