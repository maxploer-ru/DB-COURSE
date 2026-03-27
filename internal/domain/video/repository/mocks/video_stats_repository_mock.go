package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type VideoStatsRepositoryMock struct {
	mock.Mock
}

func (m *VideoStatsRepositoryMock) GetViewsCount(ctx context.Context, videoID int) (int, error) {
	args := m.Called(ctx, videoID)
	return args.Int(0), args.Error(1)
}

func (m *VideoStatsRepositoryMock) GetLikesDislikes(ctx context.Context, videoID int) (int, int, error) {
	args := m.Called(ctx, videoID)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *VideoStatsRepositoryMock) GetCommentsCount(ctx context.Context, videoID int) (int, error) {
	args := m.Called(ctx, videoID)
	return args.Int(0), args.Error(1)
}
