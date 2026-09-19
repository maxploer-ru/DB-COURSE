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

type VideoInteractionServiceTestSuite struct {
	suite.Suite
	mockRatingRepo  *mocks.VideoRatingRepository
	mockViewingRepo *mocks.ViewingRepository
	mockVideoRepo   *mocks.VideoRepository
	mockCommentRepo *mocks.CommentRepository
	mockStatsCache  *mocks.VideoStatsCache
	service         service.VideoInteractionService
	ratingMother    mother.VideoInteractionMother
	videoMother     mother.VideoMother
}

func (s *VideoInteractionServiceTestSuite) SetupTest() {
	s.mockRatingRepo = mocks.NewVideoRatingRepository(s.T())
	s.mockViewingRepo = mocks.NewViewingRepository(s.T())
	s.mockVideoRepo = mocks.NewVideoRepository(s.T())
	s.mockCommentRepo = mocks.NewCommentRepository(s.T())
	s.mockStatsCache = mocks.NewVideoStatsCache(s.T())

	s.service = service.NewVideoInteractionService(
		s.mockRatingRepo,
		s.mockViewingRepo,
		s.mockVideoRepo,
		s.mockCommentRepo,
		s.mockStatsCache,
	)
	s.ratingMother = mother.VideoInteractionMother{}
	s.videoMother = mother.VideoMother{}
}

func (s *VideoInteractionServiceTestSuite) TestRate_Positive_NewLike_StateTransition() {
	ctx := context.Background()
	rating := s.ratingMother.LikedRating()

	video := s.videoMother.ReadyVideoForChannel(1)
	video.ID = rating.VideoID

	s.mockVideoRepo.On("GetByID", ctx, rating.VideoID).Return(video, nil)
	s.mockRatingRepo.On("GetByUserAndVideo", ctx, rating.UserID, rating.VideoID).Return(nil, nil)
	s.mockRatingRepo.On("Create", ctx, mock.AnythingOfType("*domain.VideoRating")).Return(nil)
	s.mockStatsCache.On("IncrLikes", ctx, rating.VideoID).Return(nil)

	err := s.service.Rate(ctx, rating.UserID, rating.VideoID, domain.RatingActionLike)

	s.NoError(err)
	s.mockRatingRepo.AssertExpectations(s.T())
}

func (s *VideoInteractionServiceTestSuite) TestRate_Negative_VideoPending_EquivalencePartitioning() {
	ctx := context.Background()
	rating := s.ratingMother.LikedRating()

	video := s.videoMother.PendingVideoForChannel(1)
	video.ID = rating.VideoID

	s.mockVideoRepo.On("GetByID", ctx, rating.VideoID).Return(video, nil)

	err := s.service.Rate(ctx, rating.UserID, rating.VideoID, domain.RatingActionLike)

	s.ErrorIs(err, domain.ErrVideoNotFound)
	s.mockRatingRepo.AssertNotCalled(s.T(), "Create")
}

func (s *VideoInteractionServiceTestSuite) TestRecordView_Positive_StateTransition() {
	ctx := context.Background()
	userID := 1
	video := s.videoMother.ReadyVideoForChannel(1)

	s.mockVideoRepo.On("GetByID", ctx, video.ID).Return(video, nil)
	s.mockViewingRepo.On("Create", ctx, mock.AnythingOfType("*domain.Viewing")).Return(nil)
	s.mockStatsCache.On("IncrViews", ctx, video.ID).Return(nil)

	err := s.service.RecordView(ctx, userID, video.ID)

	s.NoError(err)
	s.mockViewingRepo.AssertExpectations(s.T())
}

func (s *VideoInteractionServiceTestSuite) TestRecordView_Negative_VideoNotFound_EquivalencePartitioning() {
	ctx := context.Background()
	s.mockVideoRepo.On("GetByID", ctx, 999).Return(nil, nil)

	err := s.service.RecordView(ctx, 1, 999)

	s.ErrorIs(err, domain.ErrVideoNotFound)
	s.mockViewingRepo.AssertNotCalled(s.T(), "Create")
}

func (s *VideoInteractionServiceTestSuite) TestGetStats_Positive_CacheHit_EquivalencePartitioning() {
	ctx := context.Background()
	expected := &domain.VideoStats{Views: 12, Likes: 4, Dislikes: 1, Comments: 3}
	s.mockStatsCache.On("GetStats", ctx, 1).Return(expected, true, nil)

	stats, err := s.service.GetStats(ctx, 1)

	s.NoError(err)
	s.Equal(expected, stats)
}

func (s *VideoInteractionServiceTestSuite) TestGetStats_Negative_RatingRepositoryError_EquivalencePartitioning() {
	ctx := context.Background()
	s.mockStatsCache.On("GetStats", ctx, 1).Return((*domain.VideoStats)(nil), false, nil)
	s.mockRatingRepo.On("GetStats", ctx, 1).Return(0, 0, domain.ErrInternalServer)

	stats, err := s.service.GetStats(ctx, 1)

	s.ErrorIs(err, domain.ErrInternalServer)
	s.Nil(stats)
}

func (s *VideoInteractionServiceTestSuite) TestGetStatsBatch_Positive_CacheHit_EquivalencePartitioning() {
	ctx := context.Background()
	first := &domain.VideoStats{Views: 1}
	second := &domain.VideoStats{Views: 2}
	s.mockStatsCache.On("GetStats", ctx, 1).Return(first, true, nil)
	s.mockStatsCache.On("GetStats", ctx, 2).Return(second, true, nil)

	stats, err := s.service.GetStatsBatch(ctx, []int{1, 2})

	s.NoError(err)
	s.Equal(first, stats[1])
	s.Equal(second, stats[2])
}

func (s *VideoInteractionServiceTestSuite) TestGetStatsBatch_Negative_EmptyInput_BoundaryValueAnalysis() {
	ctx := context.Background()

	stats, err := s.service.GetStatsBatch(ctx, nil)

	s.NoError(err)
	s.Empty(stats)
}

func TestVideoInteractionServiceSuite(t *testing.T) {
	suite.Run(t, new(VideoInteractionServiceTestSuite))
}
