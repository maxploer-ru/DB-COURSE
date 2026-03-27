package mocks

import (
	"ZVideo/internal/domain/channel/entity"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockChannelRepository struct {
	mock.Mock
}

func (m *MockChannelRepository) Create(ctx context.Context, channel *entity.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *MockChannelRepository) GetByID(ctx context.Context, id int) (*entity.Channel, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Channel), args.Error(1)
}

func (m *MockChannelRepository) GetByUserID(ctx context.Context, userID int) ([]*entity.Channel, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Channel), args.Error(1)
}

func (m *MockChannelRepository) GetByName(ctx context.Context, name string) (*entity.Channel, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Channel), args.Error(1)
}

func (m *MockChannelRepository) Update(ctx context.Context, channel *entity.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *MockChannelRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockChannelRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}
