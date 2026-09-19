package service_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"ZVideo/internal/testing/mocks"
	"ZVideo/internal/testing/mother"
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CommentInteractionServiceTestSuite struct {
	suite.Suite
	mockRatingRepo  *mocks.CommentRatingRepository
	mockCommentRepo *mocks.CommentRepository
	mockStatsCache  *mocks.CommentStatsCache
	service         service.CommentInteractionService
	comMother       mother.CommentMother
	rateMother      mother.CommentRatingMother
}

func (s *CommentInteractionServiceTestSuite) SetupTest() {
	s.mockRatingRepo = mocks.NewCommentRatingRepository(s.T())
	s.mockCommentRepo = mocks.NewCommentRepository(s.T())
	s.mockStatsCache = mocks.NewCommentStatsCache(s.T())
	s.service = service.NewCommentInteractionService(s.mockRatingRepo, s.mockCommentRepo, s.mockStatsCache)
	s.comMother = mother.CommentMother{}
	s.rateMother = mother.CommentRatingMother{}
}

func (s *CommentInteractionServiceTestSuite) TestRate_Positive_NewLike_StateTransition() {
	ctx := context.Background()
	comment := s.comMother.CommentForVideo(1, 1)

	s.mockCommentRepo.On("GetByID", ctx, comment.ID).Return(comment, nil)
	s.mockRatingRepo.On("GetByUserAndComment", ctx, 2, comment.ID).Return(nil, nil)
	s.mockRatingRepo.On("Create", ctx, mock.AnythingOfType("*domain.CommentRating")).Return(nil)
	s.mockStatsCache.On("IncrLikes", ctx, comment.ID).Return(nil)

	err := s.service.Rate(ctx, 2, comment.ID, domain.RatingActionLike)

	s.NoError(err)
}

func (s *CommentInteractionServiceTestSuite) TestRate_Positive_ChangeLikeToDislike_StateTransition() {
	ctx := context.Background()
	comment := s.comMother.CommentForVideo(1, 1)
	existing := s.rateMother.LikeForComment(2, comment.ID)

	s.mockCommentRepo.On("GetByID", ctx, comment.ID).Return(comment, nil)
	s.mockRatingRepo.On("GetByUserAndComment", ctx, 2, comment.ID).Return(existing, nil)
	s.mockRatingRepo.On("Update", ctx, existing).Return(nil)
	s.mockStatsCache.On("IncrDislikes", ctx, comment.ID).Return(nil)
	s.mockStatsCache.On("DecrLikes", ctx, comment.ID).Return(nil)

	err := s.service.Rate(ctx, 2, comment.ID, domain.RatingActionDislike)

	s.NoError(err)
}

func (s *CommentInteractionServiceTestSuite) TestRate_Positive_RemoveRating_StateTransition() {
	ctx := context.Background()
	comment := s.comMother.CommentForVideo(1, 1)
	existing := s.rateMother.LikeForComment(2, comment.ID)

	s.mockCommentRepo.On("GetByID", ctx, comment.ID).Return(comment, nil)
	s.mockRatingRepo.On("GetByUserAndComment", ctx, 2, comment.ID).Return(existing, nil)
	s.mockRatingRepo.On("Delete", ctx, 2, comment.ID).Return(nil)
	s.mockStatsCache.On("DecrLikes", ctx, comment.ID).Return(nil)

	err := s.service.Rate(ctx, 2, comment.ID, domain.RatingActionRemove)

	s.NoError(err)
}

func (s *CommentInteractionServiceTestSuite) TestRate_Negative_CommentNotFound_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockCommentRepo.On("GetByID", ctx, 999).Return(nil, nil)

	err := s.service.Rate(ctx, 2, 999, domain.RatingActionLike)

	s.ErrorIs(err, domain.ErrCommentNotFound)
}

func (s *CommentInteractionServiceTestSuite) TestGetStats_Positive_CacheHit_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockStatsCache.On("GetStats", ctx, 1).Return(int64(10), int64(2), true, nil)

	likes, dislikes, err := s.service.GetStats(ctx, 1)

	s.NoError(err)
	s.Equal(int64(10), likes)
	s.Equal(int64(2), dislikes)
}

func (s *CommentInteractionServiceTestSuite) TestGetStats_Positive_DBFallback_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockStatsCache.On("GetStats", ctx, 1).Return(int64(0), int64(0), false, nil)
	s.mockRatingRepo.On("GetStats", ctx, 1).Return(int64(15), int64(3), nil)
	s.mockStatsCache.On("SetStats", ctx, 1, int64(15), int64(3)).Return(nil)

	likes, dislikes, err := s.service.GetStats(ctx, 1)

	s.NoError(err)
	s.Equal(int64(15), likes)
	s.Equal(int64(3), dislikes)
}

func (s *CommentInteractionServiceTestSuite) TestGetStats_Negative_RatingRepositoryError_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockStatsCache.On("GetStats", ctx, 1).Return(int64(0), int64(0), false, nil)
	s.mockRatingRepo.On("GetStats", ctx, 1).Return(int64(0), int64(0), domain.ErrInternalServer)

	likes, dislikes, err := s.service.GetStats(ctx, 1)

	s.ErrorIs(err, domain.ErrInternalServer)
	s.Zero(likes)
	s.Zero(dislikes)
}

func TestCommentInteractionServiceSuite(t *testing.T) {
	suite.Run(t, new(CommentInteractionServiceTestSuite))
}
