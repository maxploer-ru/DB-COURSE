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

func TestVideoInteractionServiceSuite(t *testing.T) {
	suite.Run(t, new(VideoInteractionServiceTestSuite))
}
