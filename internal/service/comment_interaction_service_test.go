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

func TestCommentInteractionService_Like(t *testing.T) {
	ctx := context.Background()
	ratingRepo := mocks.NewCommentRatingRepository(t)
	commentRepo := mocks.NewCommentRepository(t)
	statsCache := mocks.NewCommentStatsCache(t)

	commentRepo.On("GetByID", ctx, 1).Return(&domain.Comment{ID: 1}, nil)
	ratingRepo.On("GetByUserAndComment", ctx, 2, 1).Return((*domain.CommentRating)(nil), nil)
	ratingRepo.On("Create", ctx, mock.MatchedBy(func(r *domain.CommentRating) bool {
		return r.UserID == 2 && r.CommentID == 1 && r.Liked
	})).Return(nil)
	statsCache.On("IncrLikes", ctx, 1).Return(nil)

	svc := service.NewCommentInteractionService(ratingRepo, commentRepo, statsCache)
	err := svc.Like(ctx, 2, 1)
	require.NoError(t, err)
}

func TestCommentInteractionService_Dislike(t *testing.T) {
	ctx := context.Background()
	ratingRepo := mocks.NewCommentRatingRepository(t)
	commentRepo := mocks.NewCommentRepository(t)
	statsCache := mocks.NewCommentStatsCache(t)

	commentRepo.On("GetByID", ctx, 3).Return(&domain.Comment{ID: 3}, nil)
	ratingRepo.On("GetByUserAndComment", ctx, 4, 3).Return((*domain.CommentRating)(nil), nil)
	ratingRepo.On("Create", ctx, mock.MatchedBy(func(r *domain.CommentRating) bool {
		return r.UserID == 4 && r.CommentID == 3 && !r.Liked
	})).Return(nil)
	statsCache.On("IncrDislikes", ctx, 3).Return(nil)

	svc := service.NewCommentInteractionService(ratingRepo, commentRepo, statsCache)
	err := svc.Dislike(ctx, 4, 3)
	require.NoError(t, err)
}

func TestCommentInteractionService_RemoveRating(t *testing.T) {
	ctx := context.Background()
	ratingRepo := mocks.NewCommentRatingRepository(t)
	commentRepo := mocks.NewCommentRepository(t)
	statsCache := mocks.NewCommentStatsCache(t)

	ratingRepo.On("GetByUserAndComment", ctx, 5, 6).Return(&domain.CommentRating{UserID: 5, CommentID: 6, Liked: true}, nil)
	ratingRepo.On("Delete", ctx, 5, 6).Return(nil)
	statsCache.On("DecrLikes", ctx, 6).Return(nil)

	svc := service.NewCommentInteractionService(ratingRepo, commentRepo, statsCache)
	err := svc.RemoveRating(ctx, 5, 6)
	require.NoError(t, err)
}

func TestCommentInteractionService_GetStats(t *testing.T) {
	ctx := context.Background()
	ratingRepo := mocks.NewCommentRatingRepository(t)
	commentRepo := mocks.NewCommentRepository(t)
	statsCache := mocks.NewCommentStatsCache(t)

	statsCache.On("GetStats", ctx, 8).Return(int64(2), int64(1), true, nil)

	svc := service.NewCommentInteractionService(ratingRepo, commentRepo, statsCache)
	likes, dislikes, err := svc.GetStats(ctx, 8)
	require.NoError(t, err)
	require.Equal(t, int64(2), likes)
	require.Equal(t, int64(1), dislikes)
}
