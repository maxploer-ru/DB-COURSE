package service

import (
	"ZVideo/internal/domain/video/entity"
	"ZVideo/internal/domain/video/repository/mocks"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetVideoStats_FromCache(t *testing.T) {
	repo := new(mocks.VideoStatsRepositoryMock)
	cache := new(mocks.VideoStatsCacheMock)

	svc := NewVideoStatsService(repo, cache)

	expected := &entity.VideoStats{VideoID: 1, ViewsCount: 10}

	cache.On("Get", mock.Anything, 1).
		Return(expected, true, nil)

	stats, err := svc.GetVideoStats(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, stats)

	repo.AssertNotCalled(t, "GetViewsCount")
}

func TestGetVideoStats_FromDB(t *testing.T) {
	repo := new(mocks.VideoStatsRepositoryMock)
	cache := new(mocks.VideoStatsCacheMock)

	svc := NewVideoStatsService(repo, cache)

	cache.On("Get", mock.Anything, 1).
		Return(nil, false, nil)

	repo.On("GetViewsCount", mock.Anything, 1).Return(100, nil)
	repo.On("GetLikesDislikes", mock.Anything, 1).Return(10, 2, nil)
	repo.On("GetCommentsCount", mock.Anything, 1).Return(5, nil)

	cache.On("Set", mock.Anything, 1, mock.Anything).Return(nil)

	stats, err := svc.GetVideoStats(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, 100, stats.ViewsCount)
}

func TestGetVideoStats_ErrorFromRepo(t *testing.T) {
	repo := new(mocks.VideoStatsRepositoryMock)
	cache := new(mocks.VideoStatsCacheMock)

	svc := NewVideoStatsService(repo, cache)

	cache.On("Get", mock.Anything, 1).
		Return(nil, false, nil)

	repo.On("GetViewsCount", mock.Anything, 1).
		Return(0, errors.New("db error"))

	stats, err := svc.GetVideoStats(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, stats)
}
