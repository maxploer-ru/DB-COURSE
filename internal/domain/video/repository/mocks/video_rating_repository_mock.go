package mocks

import (
	"ZVideo/internal/domain/video/entity"
	"context"

	"github.com/stretchr/testify/mock"
)

type VideoRatingRepositoryMock struct {
	mock.Mock
}

func (m *VideoRatingRepositoryMock) Create(ctx context.Context, rating *entity.VideoRating) error {
	args := m.Called(ctx, rating)
	return args.Error(0)
}

func (m *VideoRatingRepositoryMock) Update(ctx context.Context, rating *entity.VideoRating) error {
	args := m.Called(ctx, rating)
	return args.Error(0)
}

func (m *VideoRatingRepositoryMock) Delete(ctx context.Context, userID, videoID int) error {
	args := m.Called(ctx, userID, videoID)
	return args.Error(0)
}

func (m *VideoRatingRepositoryMock) GetByUserAndVideo(ctx context.Context, userID, videoID int) (*entity.VideoRating, error) {
	args := m.Called(ctx, userID, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.VideoRating), args.Error(1)
}
