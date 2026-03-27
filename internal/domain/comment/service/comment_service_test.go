package service_test

import (
	"ZVideo/internal/domain/comment/entity"
	"ZVideo/internal/domain/comment/repository/mocks"
	"ZVideo/internal/domain/comment/service"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockVideoChecker struct {
	mock.Mock
}

func (m *mockVideoChecker) Exists(ctx context.Context, videoID int) (bool, error) {
	args := m.Called(ctx, videoID)
	return args.Bool(0), args.Error(1)
}

func TestCommentService_CreateComment_Success(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	ratingRepo := new(mocks.MockCommentRatingRepository)
	videoChecker := new(mockVideoChecker)
	svc := service.NewCommentService(commentRepo, ratingRepo, videoChecker)

	videoChecker.On("Exists", mock.Anything, 10).Return(true, nil)
	commentRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *entity.Comment) bool {
		return c.UserID == 1 && c.VideoID == 10 && c.Content == "Great video!"
	})).Return(nil)

	comment, err := svc.CreateComment(context.Background(), 1, 10, "Great video!")
	assert.NoError(t, err)
	assert.Equal(t, "Great video!", comment.Content)
	videoChecker.AssertExpectations(t)
	commentRepo.AssertExpectations(t)
}

func TestCommentService_CreateComment_EmptyContent(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	ratingRepo := new(mocks.MockCommentRatingRepository)
	videoChecker := new(mockVideoChecker)
	svc := service.NewCommentService(commentRepo, ratingRepo, videoChecker)

	_, err := svc.CreateComment(context.Background(), 1, 10, "")
	assert.Error(t, err)
	assert.Equal(t, service.ErrEmptyContent, err)
}

func TestCommentService_UpdateComment_NotOwner(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	ratingRepo := new(mocks.MockCommentRatingRepository)
	videoChecker := new(mockVideoChecker)
	svc := service.NewCommentService(commentRepo, ratingRepo, videoChecker)

	existing := &entity.Comment{ID: 5, UserID: 2, Content: "old"}
	commentRepo.On("GetByID", mock.Anything, 5).Return(existing, nil)

	_, err := svc.UpdateComment(context.Background(), 5, 1, "new")
	assert.Error(t, err)
	assert.Equal(t, service.ErrNotAuthorized, err)
}
