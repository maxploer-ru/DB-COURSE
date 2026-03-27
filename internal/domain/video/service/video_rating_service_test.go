package service

import (
	"ZVideo/internal/domain/video/entity"
	"ZVideo/internal/domain/video/repository/mocks"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRateVideo_Create(t *testing.T) {
	videoRepo := new(mocks.VideoRepositoryMock)
	ratingRepo := new(mocks.VideoRatingRepositoryMock)
	cache := new(mocks.VideoStatsCacheMock)

	svc := NewVideoRatingService(videoRepo, ratingRepo, cache)

	videoRepo.On("GetByID", mock.Anything, 1).
		Return(&entity.Video{ID: 1}, nil)

	ratingRepo.On("GetByUserAndVideo", mock.Anything, 1, 1).
		Return(nil, ErrRatingNotFound)

	ratingRepo.On("Create", mock.Anything, mock.Anything).
		Return(nil)

	cache.On("Invalidate", mock.Anything, 1).Return(nil)

	err := svc.RateVideo(context.Background(), 1, 1, true)

	assert.NoError(t, err)
}

func TestRateVideo_Update(t *testing.T) {
	videoRepo := new(mocks.VideoRepositoryMock)
	ratingRepo := new(mocks.VideoRatingRepositoryMock)
	cache := new(mocks.VideoStatsCacheMock)

	svc := NewVideoRatingService(videoRepo, ratingRepo, cache)

	existing := &entity.VideoRating{
		UserID:  1,
		VideoID: 1,
		Liked:   false,
	}

	videoRepo.On("GetByID", mock.Anything, 1).
		Return(&entity.Video{ID: 1}, nil)

	ratingRepo.On("GetByUserAndVideo", mock.Anything, 1, 1).
		Return(existing, nil)

	ratingRepo.On("Update", mock.Anything, mock.Anything).
		Return(nil)

	cache.On("Invalidate", mock.Anything, 1).Return(nil)

	err := svc.RateVideo(context.Background(), 1, 1, true)

	assert.NoError(t, err)
}

func TestRateVideo_VideoNotFound(t *testing.T) {
	videoRepo := new(mocks.VideoRepositoryMock)
	ratingRepo := new(mocks.VideoRatingRepositoryMock)
	cache := new(mocks.VideoStatsCacheMock)

	svc := NewVideoRatingService(videoRepo, ratingRepo, cache)

	videoRepo.On("GetByID", mock.Anything, 1).
		Return(nil, nil)

	err := svc.RateVideo(context.Background(), 1, 1, true)

	assert.ErrorIs(t, err, ErrVideoNotFound)
}
