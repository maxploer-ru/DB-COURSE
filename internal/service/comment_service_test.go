package service_test

import (
	"ZVideo/internal/domain"
	service "ZVideo/internal/service"
	"ZVideo/mocks"
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCommentService_Create(t *testing.T) {
	ctx := context.Background()
	commentRepo := mocks.NewCommentRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	countCache := mocks.NewVideoStatsCache(t)
	channelSvc := mocks.NewChannelService(t)

	videoRepo.On("GetByID", ctx, 1).Return(&domain.Video{ID: 1}, nil)
	commentRepo.On("Create", ctx, mock.MatchedBy(func(c *domain.Comment) bool {
		return c.UserID == 2 && c.VideoID == 1 && c.Content == "hi"
	})).Return(nil)
	countCache.On("IncrComments", ctx, 1).Return(nil)

	svc := service.NewCommentService(commentRepo, videoRepo, countCache, channelSvc)
	comment, err := svc.Create(ctx, 2, 1, "hi")
	require.NoError(t, err)
	require.Equal(t, 2, comment.UserID)
}

func TestCommentService_GetByID(t *testing.T) {
	ctx := context.Background()
	commentRepo := mocks.NewCommentRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	countCache := mocks.NewVideoStatsCache(t)
	channelSvc := mocks.NewChannelService(t)

	commentRepo.On("GetByID", ctx, 3).Return(&domain.Comment{ID: 3}, nil)

	svc := service.NewCommentService(commentRepo, videoRepo, countCache, channelSvc)
	comment, err := svc.GetByID(ctx, 3)
	require.NoError(t, err)
	require.Equal(t, 3, comment.ID)
}

func TestCommentService_ListByVideo(t *testing.T) {
	ctx := context.Background()
	commentRepo := mocks.NewCommentRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	countCache := mocks.NewVideoStatsCache(t)
	channelSvc := mocks.NewChannelService(t)

	videoRepo.On("GetByID", ctx, 4).Return(&domain.Video{ID: 4}, nil)
	commentRepo.On("ListByVideo", ctx, 4, 10, 0).Return([]*domain.Comment{{ID: 1}}, nil)

	svc := service.NewCommentService(commentRepo, videoRepo, countCache, channelSvc)
	comments, err := svc.ListByVideo(ctx, 4, 10, 0)
	require.NoError(t, err)
	require.Len(t, comments, 1)
}

func TestCommentService_Update(t *testing.T) {
	ctx := context.Background()
	commentRepo := mocks.NewCommentRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	countCache := mocks.NewVideoStatsCache(t)
	channelSvc := mocks.NewChannelService(t)

	comment := &domain.Comment{ID: 5, UserID: 2, Content: "old"}
	commentRepo.On("GetByID", ctx, 5).Return(comment, nil)
	commentRepo.On("Update", ctx, comment).Return(nil)

	svc := service.NewCommentService(commentRepo, videoRepo, countCache, channelSvc)
	updated, err := svc.Update(ctx, 5, 2, "new")
	require.NoError(t, err)
	require.Equal(t, "new", updated.Content)
}

func TestCommentService_Delete(t *testing.T) {
	ctx := context.Background()
	commentRepo := mocks.NewCommentRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	countCache := mocks.NewVideoStatsCache(t)
	channelSvc := mocks.NewChannelService(t)

	comment := &domain.Comment{ID: 6, UserID: 2, VideoID: 4}
	commentRepo.On("GetByID", ctx, 6).Return(comment, nil)
	commentRepo.On("Delete", ctx, 6).Return(nil)
	countCache.On("DecrComments", ctx, 4).Return(nil)

	svc := service.NewCommentService(commentRepo, videoRepo, countCache, channelSvc)
	err := svc.Delete(ctx, 6, 2, "user")
	require.NoError(t, err)
}

func TestCommentService_GetCount(t *testing.T) {
	ctx := context.Background()
	commentRepo := mocks.NewCommentRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	countCache := mocks.NewVideoStatsCache(t)
	channelSvc := mocks.NewChannelService(t)

	countCache.On("GetCommentsCount", ctx, 7).Return(int64(3), true, nil)

	svc := service.NewCommentService(commentRepo, videoRepo, countCache, channelSvc)
	count, err := svc.GetCount(ctx, 7)
	require.NoError(t, err)
	require.Equal(t, int64(3), count)
}
