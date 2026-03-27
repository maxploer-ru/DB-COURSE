package mocks

import (
	"ZVideo/internal/domain/video/entity"
	"context"

	"github.com/stretchr/testify/mock"
)

type VideoRepositoryMock struct {
	mock.Mock
}

func (m *VideoRepositoryMock) Create(ctx context.Context, video *entity.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *VideoRepositoryMock) GetByID(ctx context.Context, id int) (*entity.Video, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Video), args.Error(1)
}

func (m *VideoRepositoryMock) GetByChannelID(ctx context.Context, channelID int, limit, offset int) ([]*entity.Video, error) {
	args := m.Called(ctx, channelID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Video), args.Error(1)
}

func (m *VideoRepositoryMock) Update(ctx context.Context, video *entity.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *VideoRepositoryMock) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *VideoRepositoryMock) Search(ctx context.Context, query string, limit, offset int) ([]*entity.Video, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Video), args.Error(1)
}
