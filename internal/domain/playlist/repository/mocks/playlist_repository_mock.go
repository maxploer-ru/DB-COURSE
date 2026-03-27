package mocks

import (
	"ZVideo/internal/domain/playlist/entity"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockPlaylistRepository struct {
	mock.Mock
}

func (m *MockPlaylistRepository) Create(ctx context.Context, pl *entity.Playlist) error {
	args := m.Called(ctx, pl)
	return args.Error(0)
}

func (m *MockPlaylistRepository) GetByID(ctx context.Context, id int) (*entity.Playlist, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Playlist), args.Error(1)
}

func (m *MockPlaylistRepository) GetByChannelID(ctx context.Context, channelID int, limit, offset int) ([]*entity.Playlist, error) {
	args := m.Called(ctx, channelID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Playlist), args.Error(1)
}

func (m *MockPlaylistRepository) Update(ctx context.Context, pl *entity.Playlist) error {
	args := m.Called(ctx, pl)
	return args.Error(0)
}

func (m *MockPlaylistRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
