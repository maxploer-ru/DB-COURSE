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

func TestVideoInteractionService_Like(t *testing.T) {
	ctx := context.Background()
	ratingRepo := mocks.NewVideoRatingRepository(t)
	viewingRepo := mocks.NewViewingRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	commentRepo := mocks.NewCommentRepository(t)
	statsCache := mocks.NewVideoStatsCache(t)

	videoRepo.On("GetByID", ctx, 1).Return(&domain.Video{ID: 1}, nil)
	ratingRepo.On("GetByUserAndVideo", ctx, 2, 1).Return((*domain.VideoRating)(nil), nil)
	ratingRepo.On("Create", ctx, mock.MatchedBy(func(r *domain.VideoRating) bool {
		return r.UserID == 2 && r.VideoID == 1 && r.Liked
	})).Return(nil)
	statsCache.On("IncrLikes", ctx, 1).Return(nil)

	svc := service.NewVideoInteractionService(ratingRepo, viewingRepo, videoRepo, commentRepo, statsCache)
	err := svc.Like(ctx, 2, 1)
	require.NoError(t, err)
}

func TestVideoInteractionService_Dislike(t *testing.T) {
	ctx := context.Background()
	ratingRepo := mocks.NewVideoRatingRepository(t)
	viewingRepo := mocks.NewViewingRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	commentRepo := mocks.NewCommentRepository(t)
	statsCache := mocks.NewVideoStatsCache(t)

	videoRepo.On("GetByID", ctx, 3).Return(&domain.Video{ID: 3}, nil)
	ratingRepo.On("GetByUserAndVideo", ctx, 4, 3).Return((*domain.VideoRating)(nil), nil)
	ratingRepo.On("Create", ctx, mock.MatchedBy(func(r *domain.VideoRating) bool {
		return r.UserID == 4 && r.VideoID == 3 && !r.Liked
	})).Return(nil)
	statsCache.On("IncrDislikes", ctx, 3).Return(nil)

	svc := service.NewVideoInteractionService(ratingRepo, viewingRepo, videoRepo, commentRepo, statsCache)
	err := svc.Dislike(ctx, 4, 3)
	require.NoError(t, err)
}

func TestVideoInteractionService_RemoveRating(t *testing.T) {
	ctx := context.Background()
	ratingRepo := mocks.NewVideoRatingRepository(t)
	viewingRepo := mocks.NewViewingRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	commentRepo := mocks.NewCommentRepository(t)
	statsCache := mocks.NewVideoStatsCache(t)

	ratingRepo.On("GetByUserAndVideo", ctx, 5, 6).Return(&domain.VideoRating{UserID: 5, VideoID: 6, Liked: false}, nil)
	ratingRepo.On("Delete", ctx, 5, 6).Return(nil)
	statsCache.On("DecrDislikes", ctx, 6).Return(nil)

	svc := service.NewVideoInteractionService(ratingRepo, viewingRepo, videoRepo, commentRepo, statsCache)
	err := svc.RemoveRating(ctx, 5, 6)
	require.NoError(t, err)
}

func TestVideoInteractionService_RecordView(t *testing.T) {
	ctx := context.Background()
	ratingRepo := mocks.NewVideoRatingRepository(t)
	viewingRepo := mocks.NewViewingRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	commentRepo := mocks.NewCommentRepository(t)
	statsCache := mocks.NewVideoStatsCache(t)

	viewingRepo.On("Create", ctx, mock.MatchedBy(func(v *domain.Viewing) bool {
		return v.UserID == 7 && v.VideoID == 8
	})).Return(nil)
	statsCache.On("IncrViews", ctx, 8).Return(nil)

	svc := service.NewVideoInteractionService(ratingRepo, viewingRepo, videoRepo, commentRepo, statsCache)
	err := svc.RecordView(ctx, 7, 8)
	require.NoError(t, err)
}

func TestVideoInteractionService_GetStats(t *testing.T) {
	ctx := context.Background()
	ratingRepo := mocks.NewVideoRatingRepository(t)
	viewingRepo := mocks.NewViewingRepository(t)
	videoRepo := mocks.NewVideoRepository(t)
	commentRepo := mocks.NewCommentRepository(t)
	statsCache := mocks.NewVideoStatsCache(t)

	statsCache.On("GetStats", ctx, 9).Return(&domain.VideoStats{Views: 1, Likes: 2, Dislikes: 0, Comments: 3}, true, nil)

	svc := service.NewVideoInteractionService(ratingRepo, viewingRepo, videoRepo, commentRepo, statsCache)
	stats, err := svc.GetStats(ctx, 9)
	require.NoError(t, err)
	require.Equal(t, 1, stats.Views)
}
