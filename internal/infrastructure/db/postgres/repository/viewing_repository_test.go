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

func (s *ViewingRepositoryTestSuite) TestCreateAndGetTotalViews_Positive_StateTransition() {
	ctx := context.Background()
	view1 := s.mother.ValidViewing()
	view1.UserID = s.testUser.ID
	view1.VideoID = s.testVideo.ID

	view2 := s.mother.ValidViewing()
	view2.UserID = s.testUser.ID
	view2.VideoID = s.testVideo.ID

	err1 := s.repo.Create(ctx, view1)
	err2 := s.repo.Create(ctx, view2)

	s.NoError(err1)
	s.NoError(err2)

	totalViews, err := s.repo.GetTotalViews(ctx, s.testVideo.ID)

	s.NoError(err)
	s.Equal(2, totalViews)
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
