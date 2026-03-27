package service_test

import (
	"ZVideo/internal/domain/playlist/entity"
	"ZVideo/internal/domain/playlist/repository/mocks"
	"ZVideo/internal/domain/playlist/service"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockChannelChecker struct {
	mock.Mock
}

func (m *mockChannelChecker) IsOwner(ctx context.Context, channelID, userID int) (bool, error) {
	args := m.Called(ctx, channelID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *mockChannelChecker) Exists(ctx context.Context, channelID int) (bool, error) {
	args := m.Called(ctx, channelID)
	return args.Bool(0), args.Error(1)
}

type mockVideoChecker struct {
	mock.Mock
}

func (m *mockVideoChecker) Exists(ctx context.Context, videoID int) (bool, error) {
	args := m.Called(ctx, videoID)
	return args.Bool(0), args.Error(1)
}

func TestPlaylistService_CreatePlaylist_Success(t *testing.T) {
	playlistRepo := new(mocks.MockPlaylistRepository)
	playlistItemRepo := new(mocks.MockPlaylistItemRepository)
	channelChecker := new(mockChannelChecker)
	videoChecker := new(mockVideoChecker)
	svc := service.NewPlaylistService(playlistRepo, playlistItemRepo, channelChecker, videoChecker)

	channelChecker.On("Exists", mock.Anything, 5).Return(true, nil)
	channelChecker.On("IsOwner", mock.Anything, 5, 10).Return(true, nil)
	playlistRepo.On("Create", mock.Anything, mock.MatchedBy(func(pl *entity.Playlist) bool {
		return pl.ChannelID == 5 && pl.Name == "My Playlist"
	})).Return(nil)

	playlist, err := svc.CreatePlaylist(context.Background(), 5, 10, "My Playlist", "desc")
	assert.NoError(t, err)
	assert.Equal(t, "My Playlist", playlist.Name)
	channelChecker.AssertExpectations(t)
	playlistRepo.AssertExpectations(t)
}

func TestPlaylistService_CreatePlaylist_EmptyName(t *testing.T) {
	playlistRepo := new(mocks.MockPlaylistRepository)
	playlistItemRepo := new(mocks.MockPlaylistItemRepository)
	channelChecker := new(mockChannelChecker)
	videoChecker := new(mockVideoChecker)
	svc := service.NewPlaylistService(playlistRepo, playlistItemRepo, channelChecker, videoChecker)

	_, err := svc.CreatePlaylist(context.Background(), 5, 10, "", "desc")
	assert.Error(t, err)
	assert.Equal(t, service.ErrEmptyName, err)
}

func TestPlaylistService_AddVideoToPlaylist_Success(t *testing.T) {
	playlistRepo := new(mocks.MockPlaylistRepository)
	playlistItemRepo := new(mocks.MockPlaylistItemRepository)
	channelChecker := new(mockChannelChecker)
	videoChecker := new(mockVideoChecker)
	svc := service.NewPlaylistService(playlistRepo, playlistItemRepo, channelChecker, videoChecker)

	playlist := &entity.Playlist{ID: 1, ChannelID: 5}
	playlistRepo.On("GetByID", mock.Anything, 1).Return(playlist, nil)
	channelChecker.On("IsOwner", mock.Anything, 5, 10).Return(true, nil)
	videoChecker.On("Exists", mock.Anything, 100).Return(true, nil)
	playlistItemRepo.On("GetItem", mock.Anything, 1, 100).Return(nil, nil)
	playlistItemRepo.On("GetMaxPosition", mock.Anything, 1).Return(3, nil)
	playlistItemRepo.On("AddItem", mock.Anything, mock.Anything).Return(nil)

	err := svc.AddVideoToPlaylist(context.Background(), 1, 100, 10, nil)
	assert.NoError(t, err)
}

func TestPlaylistService_AddVideoToPlaylist_AlreadyExists(t *testing.T) {
	playlistRepo := new(mocks.MockPlaylistRepository)
	playlistItemRepo := new(mocks.MockPlaylistItemRepository)
	channelChecker := new(mockChannelChecker)
	videoChecker := new(mockVideoChecker)
	svc := service.NewPlaylistService(playlistRepo, playlistItemRepo, channelChecker, videoChecker)

	playlist := &entity.Playlist{ID: 1, ChannelID: 5}
	playlistRepo.On("GetByID", mock.Anything, 1).Return(playlist, nil)
	channelChecker.On("IsOwner", mock.Anything, 5, 10).Return(true, nil)
	videoChecker.On("Exists", mock.Anything, 100).Return(true, nil)
	playlistItemRepo.On("GetItem", mock.Anything, 1, 100).Return(&entity.PlaylistItem{}, nil)

	err := svc.AddVideoToPlaylist(context.Background(), 1, 100, 10, nil)
	assert.Error(t, err)
	assert.Equal(t, service.ErrVideoAlreadyInPlaylist, err)
}

func TestPlaylistService_RemoveVideoFromPlaylist_Success(t *testing.T) {
	playlistRepo := new(mocks.MockPlaylistRepository)
	playlistItemRepo := new(mocks.MockPlaylistItemRepository)
	channelChecker := new(mockChannelChecker)
	videoChecker := new(mockVideoChecker)
	svc := service.NewPlaylistService(playlistRepo, playlistItemRepo, channelChecker, videoChecker)

	playlist := &entity.Playlist{ID: 1, ChannelID: 5}
	playlistRepo.On("GetByID", mock.Anything, 1).Return(playlist, nil)
	channelChecker.On("IsOwner", mock.Anything, 5, 10).Return(true, nil)
	playlistItemRepo.On("RemoveItem", mock.Anything, 1, 100).Return(nil)

	err := svc.RemoveVideoFromPlaylist(context.Background(), 1, 100, 10)
	assert.NoError(t, err)
}
