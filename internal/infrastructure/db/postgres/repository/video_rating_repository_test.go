package repository_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
	"ZVideo/internal/infrastructure/db/postgres/repository"
	"ZVideo/internal/testing/db"
	"ZVideo/internal/testing/mother"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type VideoRatingRepositoryTestSuite struct {
	suite.Suite
	pgContainer *db.PostgresContainer
	db          *gorm.DB
	tx          *gorm.DB
	repo        *repository.VideoRatingRepository
	mother      mother.VideoInteractionMother
	testUser    *models.User
	testVideo   *models.Video
}

func (s *VideoRatingRepositoryTestSuite) SetupSuite() {
	s.db = sharedDB
	s.mother = mother.VideoInteractionMother{}
}

func (s *VideoRatingRepositoryTestSuite) SetupTest() {
	s.tx = s.db.Begin()
	s.repo = repository.NewVideoRatingRepository(s.tx)

	s.testUser = &models.User{Username: "rater", Email: "r@t.com", PasswordHash: "x", RoleID: 1}
	s.tx.Create(s.testUser)

	channel := &models.Channel{UserID: s.testUser.ID, Name: "Rate Chan"}
	s.tx.Create(channel)

	s.testVideo = &models.Video{ChannelID: channel.ID, Title: "Test Video", Filepath: "x", Status: "ready"}
	s.tx.Create(s.testVideo)
}

func (s *VideoRatingRepositoryTestSuite) TearDownTest() {
	s.tx.Rollback()
}

func (s *VideoRatingRepositoryTestSuite) TestCreate_Positive_StateTransition() {
	ctx := context.Background()
	rating := s.mother.LikedRating()
	rating.UserID = s.testUser.ID
	rating.VideoID = s.testVideo.ID

	err := s.repo.Create(ctx, rating)

	s.NoError(err)
}

func (s *VideoRatingRepositoryTestSuite) TestCreate_Negative_AlreadyExists_EquivalencePartitioning() {
	ctx := context.Background()
	rating := s.mother.LikedRating()
	rating.UserID = s.testUser.ID
	rating.VideoID = s.testVideo.ID
	s.NoError(s.repo.Create(ctx, rating))

	err := s.repo.Create(ctx, &domain.VideoRating{UserID: s.testUser.ID, VideoID: s.testVideo.ID, Liked: false})

	s.ErrorIs(err, domain.ErrAlreadyRated)
}

func (s *VideoRatingRepositoryTestSuite) TestGetStats_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	rating := s.mother.LikedRating()
	rating.UserID = s.testUser.ID
	rating.VideoID = s.testVideo.ID
	s.NoError(s.repo.Create(ctx, rating))

	likes, dislikes, err := s.repo.GetStats(ctx, rating.VideoID)

	s.NoError(err)
	s.Equal(1, likes)
	s.Equal(0, dislikes)
}

func (s *VideoRatingRepositoryTestSuite) TestGetStats_Negative_CancelledContext_Exception() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	likes, dislikes, err := s.repo.GetStats(ctx, s.testVideo.ID)

	s.Error(err)
	s.Zero(likes)
	s.Zero(dislikes)
}

func (s *VideoRatingRepositoryTestSuite) TestUpdate_Positive_StateTransition() {
	ctx := context.Background()
	rating := s.mother.LikedRating()
	rating.UserID = s.testUser.ID
	rating.VideoID = s.testVideo.ID
	s.NoError(s.repo.Create(ctx, rating))
	rating.Liked = false

	err := s.repo.Update(ctx, rating)
	found, getErr := s.repo.GetByUserAndVideo(ctx, rating.UserID, rating.VideoID)

	s.NoError(err)
	s.NoError(getErr)
	s.False(found.Liked)
}

func (s *VideoRatingRepositoryTestSuite) TestUpdate_Negative_NotFound_EquivalencePartitioning() {
	err := s.repo.Update(context.Background(), &domain.VideoRating{UserID: s.testUser.ID, VideoID: s.testVideo.ID, Liked: true})

	s.ErrorIs(err, domain.ErrRatingNotFound)
}

func (s *VideoRatingRepositoryTestSuite) TestUpsert_Positive_StateTransition() {
	ctx := context.Background()
	rating := s.mother.LikedRating()
	rating.UserID = s.testUser.ID
	rating.VideoID = s.testVideo.ID

	previous, created, err := s.repo.Upsert(ctx, rating)

	s.NoError(err)
	s.True(created)
	s.Nil(previous)

	rating.Liked = false
	previous, created, err = s.repo.Upsert(ctx, rating)

	s.NoError(err)
	s.False(created)
	s.True(previous.Liked)
}

func (s *VideoRatingRepositoryTestSuite) TestUpsert_Negative_CancelledContext_Exception() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rating := s.mother.LikedRating()
	rating.UserID = s.testUser.ID
	rating.VideoID = s.testVideo.ID

	previous, created, err := s.repo.Upsert(ctx, rating)

	s.Error(err)
	s.Nil(previous)
	s.False(created)
}

func (s *VideoRatingRepositoryTestSuite) TestDeleteAndGet_Positive_StateTransition() {
	ctx := context.Background()
	rating := s.mother.LikedRating()
	rating.UserID = s.testUser.ID
	rating.VideoID = s.testVideo.ID
	s.NoError(s.repo.Create(ctx, rating))

	previous, found, err := s.repo.DeleteAndGet(ctx, rating.UserID, rating.VideoID)

	s.NoError(err)
	s.True(found)
	s.NotNil(previous)
	s.True(previous.Liked)
}

func (s *VideoRatingRepositoryTestSuite) TestDeleteAndGet_Negative_NotFound_EquivalencePartitioning() {
	previous, found, err := s.repo.DeleteAndGet(context.Background(), s.testUser.ID, s.testVideo.ID)

	s.NoError(err)
	s.False(found)
	s.Nil(previous)
}

func (s *VideoRatingRepositoryTestSuite) TestDelete_Positive_StateTransition() {
	ctx := context.Background()
	rating := s.mother.LikedRating()
	rating.UserID = s.testUser.ID
	rating.VideoID = s.testVideo.ID
	s.NoError(s.repo.Create(ctx, rating))

	err := s.repo.Delete(ctx, rating.UserID, rating.VideoID)
	found, getErr := s.repo.GetByUserAndVideo(ctx, rating.UserID, rating.VideoID)

	s.NoError(err)
	s.NoError(getErr)
	s.Nil(found)
}

func (s *VideoRatingRepositoryTestSuite) TestDelete_Negative_NotFound_EquivalencePartitioning() {
	ctx := context.Background()

	err := s.repo.Delete(ctx, s.testUser.ID, s.testVideo.ID)

	s.ErrorIs(err, domain.ErrRatingNotFound)
}

func (s *VideoRatingRepositoryTestSuite) TestGetByUserAndVideo_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	rating := s.mother.LikedRating()
	rating.UserID = s.testUser.ID
	rating.VideoID = s.testVideo.ID
	s.NoError(s.repo.Create(ctx, rating))

	found, err := s.repo.GetByUserAndVideo(ctx, rating.UserID, rating.VideoID)

	s.NoError(err)
	s.NotNil(found)
	s.Equal(rating.UserID, found.UserID)
}

func (s *VideoRatingRepositoryTestSuite) TestGetByUserAndVideo_Negative_NotFound_EquivalencePartitioning() {
	found, err := s.repo.GetByUserAndVideo(context.Background(), s.testUser.ID, s.testVideo.ID)

	s.NoError(err)
	s.Nil(found)
}

func TestVideoRatingRepositorySuite(t *testing.T) {
	suite.Run(t, new(VideoRatingRepositoryTestSuite))
}
