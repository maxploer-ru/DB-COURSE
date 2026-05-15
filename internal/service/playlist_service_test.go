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

func TestPlaylistService_Create(t *testing.T) {
	ctx := context.Background()
	playlistRepo := mocks.NewPlaylistRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	channelSvc := mocks.NewChannelService(t)

	channelSvc.On("IsOwner", ctx, 2, 5).Return(true, nil)
	playlistRepo.On("Create", ctx, mock.MatchedBy(func(p *domain.Playlist) bool {
		return p.ChannelID == 2 && p.Name == "pl"
	})).Return(nil)

	svc := service.NewPlaylistService(playlistRepo, videoRepo, channelSvc)
	playlist, err := svc.Create(ctx, 2, 5, "pl", "d")
	require.NoError(t, err)
	require.Equal(t, "pl", playlist.Name)
}

func TestPlaylistService_GetByID(t *testing.T) {
	ctx := context.Background()
	playlistRepo := mocks.NewPlaylistRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	channelSvc := mocks.NewChannelService(t)

	playlistRepo.On("GetByID", ctx, 1).Return(&domain.Playlist{ID: 1}, nil)

	svc := service.NewPlaylistService(playlistRepo, videoRepo, channelSvc)
	playlist, err := svc.GetByID(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, 1, playlist.ID)
}

func TestPlaylistService_ListByChannel(t *testing.T) {
	ctx := context.Background()
	playlistRepo := mocks.NewPlaylistRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	channelSvc := mocks.NewChannelService(t)

	channelSvc.On("Exists", ctx, 2).Return(true, nil)
	playlistRepo.On("ListByChannel", ctx, 2, 10, 0).Return([]*domain.Playlist{{ID: 1}}, nil)

	svc := service.NewPlaylistService(playlistRepo, videoRepo, channelSvc)
	playlists, err := svc.ListByChannel(ctx, 2, 10, 0)
	require.NoError(t, err)
	require.Len(t, playlists, 1)
}

func TestPlaylistService_Update(t *testing.T) {
	ctx := context.Background()
	playlistRepo := mocks.NewPlaylistRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	channelSvc := mocks.NewChannelService(t)

	playlist := &domain.Playlist{ID: 3, ChannelID: 2, Name: "old"}
	playlistRepo.On("GetByID", ctx, 3).Return(playlist, nil)
	channelSvc.On("IsOwner", ctx, 2, 5).Return(true, nil)
	playlistRepo.On("Update", ctx, playlist).Return(nil)

	svc := service.NewPlaylistService(playlistRepo, videoRepo, channelSvc)
	newName := "new"
	updated, err := svc.Update(ctx, 3, 5, &newName, nil)
	require.NoError(t, err)
	require.Equal(t, "new", updated.Name)
}

func TestPlaylistService_Delete(t *testing.T) {
	ctx := context.Background()
	playlistRepo := mocks.NewPlaylistRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	channelSvc := mocks.NewChannelService(t)

	playlistRepo.On("GetByID", ctx, 4).Return(&domain.Playlist{ID: 4, ChannelID: 2}, nil)
	channelSvc.On("IsOwner", ctx, 2, 5).Return(true, nil)
	playlistRepo.On("Delete", ctx, 4).Return(nil)

	svc := service.NewPlaylistService(playlistRepo, videoRepo, channelSvc)
	err := svc.Delete(ctx, 4, 5)
	require.NoError(t, err)
}

func TestPlaylistService_AddVideo(t *testing.T) {
	ctx := context.Background()
	playlistRepo := mocks.NewPlaylistRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	channelSvc := mocks.NewChannelService(t)

	playlistRepo.On("GetByID", ctx, 5).Return(&domain.Playlist{ID: 5, ChannelID: 2}, nil)
	channelSvc.On("IsOwner", ctx, 2, 5).Return(true, nil)
	videoRepo.On("GetByID", ctx, 10).Return(&domain.Video{ID: 10, ChannelID: 2}, nil)
	playlistRepo.On("AddVideo", ctx, 5, 10).Return(nil)

	svc := service.NewPlaylistService(playlistRepo, videoRepo, channelSvc)
	err := svc.AddVideo(ctx, 5, 10, 5)
	require.NoError(t, err)
}

func TestPlaylistService_RemoveVideo(t *testing.T) {
	ctx := context.Background()
	playlistRepo := mocks.NewPlaylistRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	channelSvc := mocks.NewChannelService(t)

	playlistRepo.On("GetByID", ctx, 6).Return(&domain.Playlist{ID: 6, ChannelID: 2}, nil)
	channelSvc.On("IsOwner", ctx, 2, 5).Return(true, nil)
	playlistRepo.On("RemoveVideo", ctx, 6, 10).Return(nil)

	svc := service.NewPlaylistService(playlistRepo, videoRepo, channelSvc)
	err := svc.RemoveVideo(ctx, 6, 10, 5)
	require.NoError(t, err)
}
