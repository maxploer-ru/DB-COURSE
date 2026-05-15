package service_test

import (
	"ZVideo/internal/domain"
	service "ZVideo/internal/service"
	"ZVideo/mocks"
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestChannelService_CreateChannel(t *testing.T) {
	ctx := context.Background()
	channelRepo := mocks.NewChannelRepository(t)
	videoRepo := mocks.NewChannelVideoFilepathRepository(t)
	storageSvc := mocks.NewStorageService(t)

	channelRepo.On("GetByUserID", ctx, 1).Return((*domain.Channel)(nil), nil)
	channelRepo.On("ExistsByName", ctx, "name").Return(false, nil)
	channelRepo.On("Create", ctx, mock.MatchedBy(func(ch *domain.Channel) bool {
		return ch.UserID == 1 && ch.Name == "name"
	})).Return(nil)

	svc := service.NewChannelService(channelRepo, videoRepo, storageSvc)
	ch, err := svc.CreateChannel(ctx, 1, "name", "desc")
	require.NoError(t, err)
	require.Equal(t, 1, ch.UserID)
	channelRepo.AssertExpectations(t)
}

func TestChannelService_GetChannel(t *testing.T) {
	ctx := context.Background()
	channelRepo := mocks.NewChannelRepository(t)
	videoRepo := mocks.NewChannelVideoFilepathRepository(t)
	storageSvc := mocks.NewStorageService(t)

	channelRepo.On("GetByID", ctx, 2).Return(&domain.Channel{ID: 2}, nil)

	svc := service.NewChannelService(channelRepo, videoRepo, storageSvc)
	ch, err := svc.GetChannel(ctx, 2)
	require.NoError(t, err)
	require.Equal(t, 2, ch.ID)
}

func TestChannelService_GetChannelByName(t *testing.T) {
	ctx := context.Background()
	channelRepo := mocks.NewChannelRepository(t)
	videoRepo := mocks.NewChannelVideoFilepathRepository(t)
	storageSvc := mocks.NewStorageService(t)

	channelRepo.On("GetByName", ctx, "n").Return(&domain.Channel{ID: 3, Name: "n"}, nil)

	svc := service.NewChannelService(channelRepo, videoRepo, storageSvc)
	ch, err := svc.GetChannelByName(ctx, "n")
	require.NoError(t, err)
	require.Equal(t, "n", ch.Name)
}

func TestChannelService_GetChannelByUserID(t *testing.T) {
	ctx := context.Background()
	channelRepo := mocks.NewChannelRepository(t)
	videoRepo := mocks.NewChannelVideoFilepathRepository(t)
	storageSvc := mocks.NewStorageService(t)

	channelRepo.On("GetByUserID", ctx, 10).Return(&domain.Channel{ID: 4, UserID: 10}, nil)

	svc := service.NewChannelService(channelRepo, videoRepo, storageSvc)
	ch, err := svc.GetChannelByUserID(ctx, 10)
	require.NoError(t, err)
	require.Equal(t, 10, ch.UserID)
}

func TestChannelService_UpdateChannel(t *testing.T) {
	ctx := context.Background()
	channelRepo := mocks.NewChannelRepository(t)
	videoRepo := mocks.NewChannelVideoFilepathRepository(t)
	storageSvc := mocks.NewStorageService(t)

	name := "new"
	ch := &domain.Channel{ID: 5, UserID: 11, Name: "old"}
	channelRepo.On("GetByID", ctx, 5).Return(ch, nil)
	channelRepo.On("ExistsByName", ctx, "new").Return(false, nil)
	channelRepo.On("Update", ctx, ch).Return(nil)

	svc := service.NewChannelService(channelRepo, videoRepo, storageSvc)
	updated, err := svc.UpdateChannel(ctx, 5, 11, &name, nil)
	require.NoError(t, err)
	require.Equal(t, "new", updated.Name)
}

func TestChannelService_DeleteChannel(t *testing.T) {
	ctx := context.Background()
	channelRepo := mocks.NewChannelRepository(t)
	videoRepo := mocks.NewChannelVideoFilepathRepository(t)
	storageSvc := mocks.NewStorageService(t)

	channelRepo.On("GetByID", ctx, 6).Return(&domain.Channel{ID: 6, UserID: 1}, nil)
	videoRepo.On("ListFilepathsByChannel", ctx, 6).Return([]string{}, nil)
	channelRepo.On("Delete", ctx, 6).Return(nil)

	svc := service.NewChannelService(channelRepo, videoRepo, storageSvc)
	err := svc.DeleteChannel(ctx, 6, 1)
	require.NoError(t, err)
}

func TestChannelService_Exists(t *testing.T) {
	ctx := context.Background()
	channelRepo := mocks.NewChannelRepository(t)
	videoRepo := mocks.NewChannelVideoFilepathRepository(t)
	storageSvc := mocks.NewStorageService(t)

	channelRepo.On("GetByID", ctx, 7).Return(&domain.Channel{ID: 7}, nil)

	svc := service.NewChannelService(channelRepo, videoRepo, storageSvc)
	ok, err := svc.Exists(ctx, 7)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestChannelService_IsOwner(t *testing.T) {
	ctx := context.Background()
	channelRepo := mocks.NewChannelRepository(t)
	videoRepo := mocks.NewChannelVideoFilepathRepository(t)
	storageSvc := mocks.NewStorageService(t)

	channelRepo.On("GetByID", ctx, 8).Return(&domain.Channel{ID: 8, UserID: 2}, nil)

	svc := service.NewChannelService(channelRepo, videoRepo, storageSvc)
	ok, err := svc.IsOwner(ctx, 8, 2)
	require.NoError(t, err)
	require.True(t, ok)
}
