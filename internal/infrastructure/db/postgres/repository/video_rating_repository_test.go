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

func (s *VideoRatingRepositoryTestSuite) TestCreateAndGetStats_Positive() {
	ctx := context.Background()
	rating := s.mother.LikedRating()
	rating.UserID = s.testUser.ID
	rating.VideoID = s.testVideo.ID

	err := s.repo.Create(ctx, rating)
	s.NoError(err)

	likes, dislikes, err := s.repo.GetStats(ctx, rating.VideoID)

	s.NoError(err)
	s.Equal(1, likes)
	s.Equal(0, dislikes)
}

func (s *VideoRatingRepositoryTestSuite) TestDelete_Negative_NotFound() {
	ctx := context.Background()

	err := s.repo.Delete(ctx, s.testUser.ID, s.testVideo.ID)

	s.ErrorIs(err, domain.ErrRatingNotFound)
}

func TestVideoRatingRepositorySuite(t *testing.T) {
	suite.Run(t, new(VideoRatingRepositoryTestSuite))
}
