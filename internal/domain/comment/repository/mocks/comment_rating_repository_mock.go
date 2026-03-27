package mocks

import (
	"ZVideo/internal/domain/comment/entity"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockCommentRatingRepository struct {
	mock.Mock
}

func (m *MockCommentRatingRepository) Create(ctx context.Context, rating *entity.CommentRating) error {
	args := m.Called(ctx, rating)
	return args.Error(0)
}

func (m *MockCommentRatingRepository) Update(ctx context.Context, rating *entity.CommentRating) error {
	args := m.Called(ctx, rating)
	return args.Error(0)
}

func (m *MockCommentRatingRepository) Delete(ctx context.Context, userID, commentID int) error {
	args := m.Called(ctx, userID, commentID)
	return args.Error(0)
}

func (m *MockCommentRatingRepository) GetByUserAndComment(ctx context.Context, userID, commentID int) (*entity.CommentRating, error) {
	args := m.Called(ctx, userID, commentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CommentRating), args.Error(1)
}

func (m *MockCommentRatingRepository) GetCommentRatingStats(ctx context.Context, commentID int) (likes, dislikes int, err error) {
	args := m.Called(ctx, commentID)
	return args.Int(0), args.Int(1), args.Error(2)
}
