package mocks

import (
	"ZVideo/internal/domain/video/entity"
	"context"

	"github.com/stretchr/testify/mock"
)

type VideoStatsCacheMock struct {
	mock.Mock
}

func (m *VideoStatsCacheMock) Get(ctx context.Context, videoID int) (*entity.VideoStats, bool, error) {
	args := m.Called(ctx, videoID)

	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*entity.VideoStats), args.Bool(1), args.Error(2)
}

func (m *VideoStatsCacheMock) Set(ctx context.Context, videoID int, stats *entity.VideoStats) error {
	return m.Called(ctx, videoID, stats).Error(0)
}

func (m *VideoStatsCacheMock) Invalidate(ctx context.Context, videoID int) error {
	return m.Called(ctx, videoID).Error(0)
}
