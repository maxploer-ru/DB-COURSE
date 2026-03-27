package mocks

import (
	"ZVideo/internal/domain/playlist/entity"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockPlaylistItemRepository struct {
	mock.Mock
}

func (m *MockPlaylistItemRepository) AddItem(ctx context.Context, item *entity.PlaylistItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockPlaylistItemRepository) RemoveItem(ctx context.Context, playlistID, videoID int) error {
	args := m.Called(ctx, playlistID, videoID)
	return args.Error(0)
}

func (m *MockPlaylistItemRepository) GetItemsByPlaylist(ctx context.Context, playlistID int) ([]*entity.PlaylistItem, error) {
	args := m.Called(ctx, playlistID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.PlaylistItem), args.Error(1)
}

func (m *MockPlaylistItemRepository) GetItem(ctx context.Context, playlistID, videoID int) (*entity.PlaylistItem, error) {
	args := m.Called(ctx, playlistID, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PlaylistItem), args.Error(1)
}

func (m *MockPlaylistItemRepository) ReorderItems(ctx context.Context, playlistID int, videoIDs []int) error {
	args := m.Called(ctx, playlistID, videoIDs)
	return args.Error(0)
}

func (m *MockPlaylistItemRepository) GetMaxPosition(ctx context.Context, playlistID int) (int, error) {
	args := m.Called(ctx, playlistID)
	return args.Int(0), args.Error(1)
}
