package mocks

import (
	"ZVideo/internal/domain/comment/entity"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) Create(ctx context.Context, comment *entity.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepository) GetByID(ctx context.Context, id int) (*entity.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Comment), args.Error(1)
}

func (m *MockCommentRepository) GetByVideoID(ctx context.Context, videoID int, limit, offset int) ([]*entity.Comment, error) {
	args := m.Called(ctx, videoID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Comment), args.Error(1)
}

func (m *MockCommentRepository) Update(ctx context.Context, comment *entity.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCommentRepository) GetCountByVideo(ctx context.Context, videoID int) (int, error) {
	args := m.Called(ctx, videoID)
	return args.Int(0), args.Error(1)
}
