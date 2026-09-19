package repository_test

import (
	"ZVideo/internal/infrastructure/db/postgres/models"
	"ZVideo/internal/infrastructure/db/postgres/repository"
	"ZVideo/internal/testing/db"
	"ZVideo/internal/testing/mother"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ViewingRepositoryTestSuite struct {
	suite.Suite
	pgContainer *db.PostgresContainer
	db          *gorm.DB
	tx          *gorm.DB
	repo        *repository.ViewingRepository
	mother      mother.ViewingMother
	testUser    *models.User
	testVideo   *models.Video
}

func (s *ViewingRepositoryTestSuite) SetupSuite() {
	s.db = sharedDB
	s.mother = mother.ViewingMother{}
}

func (s *ViewingRepositoryTestSuite) SetupTest() {
	s.tx = s.db.Begin()
	s.repo = repository.NewViewingRepository(s.tx)

	s.testUser = &models.User{Username: "viewer", Email: "v@t.com", PasswordHash: "x", RoleID: 1}
	s.tx.Create(s.testUser)

	channel := &models.Channel{UserID: s.testUser.ID, Name: "View Chan"}
	s.tx.Create(channel)

	s.testVideo = &models.Video{ChannelID: channel.ID, Title: "Test Video", Filepath: "x", Status: "ready"}
	s.tx.Create(s.testVideo)
}

func (s *ViewingRepositoryTestSuite) TearDownTest() {
	s.tx.Rollback()
}

func (s *ViewingRepositoryTestSuite) TestCreate_Positive_StateTransition() {
	ctx := context.Background()
	view := s.mother.ValidViewing()
	view.UserID = s.testUser.ID
	view.VideoID = s.testVideo.ID

	err := s.repo.Create(ctx, view)

	s.NoError(err)
}

func (s *ViewingRepositoryTestSuite) TestCreate_Negative_CancelledContext_Exception() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	view := s.mother.ValidViewing()
	view.UserID = s.testUser.ID
	view.VideoID = s.testVideo.ID

	err := s.repo.Create(ctx, view)

	s.Error(err)
}

func (s *ViewingRepositoryTestSuite) TestGetTotalViews_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	view := s.mother.ValidViewing()
	view.UserID = s.testUser.ID
	view.VideoID = s.testVideo.ID
	s.NoError(s.repo.Create(ctx, view))

	totalViews, err := s.repo.GetTotalViews(ctx, s.testVideo.ID)

	s.NoError(err)
	s.Equal(1, totalViews)
}

func (s *ViewingRepositoryTestSuite) TestGetTotalViews_Negative_NoViews_EquivalencePartitioning() {
	ctx := context.Background()

	totalViews, err := s.repo.GetTotalViews(ctx, 99999)

	s.NoError(err)
	s.Equal(0, totalViews)
}

func TestViewingRepositorySuite(t *testing.T) {
	suite.Run(t, new(ViewingRepositoryTestSuite))
}
