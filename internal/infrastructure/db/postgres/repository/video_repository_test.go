package repository_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/repository"
	"ZVideo/internal/testing/db"
	"ZVideo/internal/testing/mother"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type VideoRepositoryTestSuite struct {
	suite.Suite
	pgContainer *db.PostgresContainer
	db          *gorm.DB
	tx          *gorm.DB
	repo        *repository.VideoRepository
	userRepo    *repository.UserRepository
	chanRepo    *repository.ChannelRepository
	userMother  mother.UserMother
	chanMother  mother.ChannelMother
	vidMother   mother.VideoMother
	testUser    *domain.User
	testChannel *domain.Channel
}

func (s *VideoRepositoryTestSuite) SetupSuite() {
	s.db = sharedDB
}

func (s *VideoRepositoryTestSuite) SetupTest() {
	s.tx = s.db.Begin()
	s.repo = repository.NewVideoRepository(s.tx)
	s.userRepo = repository.NewUserRepository(s.tx)
	s.chanRepo = repository.NewChannelRepository(s.tx)
	s.userMother = mother.UserMother{}
	s.chanMother = mother.ChannelMother{}
	s.vidMother = mother.VideoMother{}

	u := s.userMother.ValidActiveUser()
	u.ID = 0
	err := s.userRepo.Create(context.Background(), u)
	s.Require().NoError(err)
	s.testUser = u

	c := s.chanMother.ChannelForUser(s.testUser.ID)
	c.ID = 0
	err = s.chanRepo.Create(context.Background(), c)
	s.Require().NoError(err)
	s.testChannel = c
}

func (s *VideoRepositoryTestSuite) TearDownTest() {
	s.tx.Rollback()
}

func (s *VideoRepositoryTestSuite) TestCreate_Positive_StateTransition() {
	ctx := context.Background()
	vid := s.vidMother.ReadyVideoForChannel(s.testChannel.ID)
	vid.ID = 0

	err := s.repo.Create(ctx, vid)

	s.NoError(err)
	s.NotZero(vid.ID)
}

func (s *VideoRepositoryTestSuite) TestCreate_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	vid := s.vidMother.ReadyVideoForChannel(999999)
	vid.ID = 0

	err := s.repo.Create(ctx, vid)

	s.Error(err)
}

func (s *VideoRepositoryTestSuite) TestGetByID_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	vid := s.vidMother.ReadyVideoForChannel(s.testChannel.ID)
	vid.ID = 0
	_ = s.repo.Create(ctx, vid)

	found, err := s.repo.GetByID(ctx, vid.ID)

	s.NoError(err)
	s.NotNil(found)
	s.Equal(vid.Title, found.Title)
	s.Equal(s.testChannel.Name, found.ChannelName)
}

func (s *VideoRepositoryTestSuite) TestGetByID_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	found, err := s.repo.GetByID(ctx, 999)

	s.NoError(err)
	s.Nil(found)
}

func (s *VideoRepositoryTestSuite) TestUpdate_Positive_StateTransition() {
	ctx := context.Background()
	vid := s.vidMother.ReadyVideoForChannel(s.testChannel.ID)
	vid.ID = 0
	_ = s.repo.Create(ctx, vid)
	vid.Title = "Updated Title"

	err := s.repo.Update(ctx, vid)

	s.NoError(err)
	found, _ := s.repo.GetByID(ctx, vid.ID)
	s.Equal("Updated Title", found.Title)
}

func (s *VideoRepositoryTestSuite) TestUpdate_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	vid := s.vidMother.ReadyVideoForChannel(s.testChannel.ID)
	vid.ID = 999999

	err := s.repo.Update(ctx, vid)

	s.ErrorIs(err, domain.ErrVideoNotFound)
}

func (s *VideoRepositoryTestSuite) TestDelete_Positive_StateTransition() {
	ctx := context.Background()
	vid := s.vidMother.ReadyVideoForChannel(s.testChannel.ID)
	vid.ID = 0
	_ = s.repo.Create(ctx, vid)

	err := s.repo.Delete(ctx, vid.ID)

	s.NoError(err)
	found, _ := s.repo.GetByID(ctx, vid.ID)
	s.Nil(found)
}

func (s *VideoRepositoryTestSuite) TestDelete_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	err := s.repo.Delete(ctx, 999)

	s.ErrorIs(err, domain.ErrVideoNotFound)
}

func (s *VideoRepositoryTestSuite) TestList_Positive_BoundaryValueAnalysis() {
	ctx := context.Background()
	vid := s.vidMother.ReadyVideoForChannel(s.testChannel.ID)
	vid.ID = 0
	_ = s.repo.Create(ctx, vid)

	list, err := s.repo.List(ctx, 10, 0)

	s.NoError(err)
	s.Len(list, 1)
}

func (s *VideoRepositoryTestSuite) TestList_Negative_BoundaryValueAnalysis() {
	ctx := context.Background()

	list, err := s.repo.List(ctx, 10, 999)

	s.NoError(err)
	s.Len(list, 0)
}

func (s *VideoRepositoryTestSuite) TestListByChannel_Positive_BoundaryValueAnalysis() {
	ctx := context.Background()
	vid := s.vidMother.ReadyVideoForChannel(s.testChannel.ID)
	vid.ID = 0
	_ = s.repo.Create(ctx, vid)

	list, err := s.repo.ListByChannel(ctx, s.testChannel.ID, 10, 0)

	s.NoError(err)
	s.Len(list, 1)
}

func (s *VideoRepositoryTestSuite) TestListByChannel_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	list, err := s.repo.ListByChannel(ctx, 999, 10, 0)

	s.NoError(err)
	s.Len(list, 0)
}

func (s *VideoRepositoryTestSuite) TestCount_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	video := s.vidMother.ReadyVideoForChannel(s.testChannel.ID)
	video.ID = 0
	s.Require().NoError(s.repo.Create(ctx, video))

	count, err := s.repo.Count(ctx)

	s.NoError(err)
	s.GreaterOrEqual(count, int64(1))
}

func (s *VideoRepositoryTestSuite) TestCount_Negative_CancelledContext_Exception() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	count, err := s.repo.Count(ctx)

	s.Error(err)
	s.Zero(count)
}

func (s *VideoRepositoryTestSuite) TestCountByChannel_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	video := s.vidMother.ReadyVideoForChannel(s.testChannel.ID)
	video.ID = 0
	s.Require().NoError(s.repo.Create(ctx, video))

	count, err := s.repo.CountByChannel(ctx, s.testChannel.ID)

	s.NoError(err)
	s.GreaterOrEqual(count, int64(1))
}

func (s *VideoRepositoryTestSuite) TestCountByChannel_Negative_CancelledContext_Exception() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	count, err := s.repo.CountByChannel(ctx, s.testChannel.ID)

	s.Error(err)
	s.Zero(count)
}

func (s *VideoRepositoryTestSuite) TestListFilepathsByChannel_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	video := s.vidMother.ReadyVideoForChannel(s.testChannel.ID)
	video.ID = 0
	s.Require().NoError(s.repo.Create(ctx, video))

	filepaths, err := s.repo.ListFilepathsByChannel(ctx, s.testChannel.ID)

	s.NoError(err)
	s.Equal([]string{video.Filepath}, filepaths)
}

func (s *VideoRepositoryTestSuite) TestListFilepathsByChannel_Negative_EquivalencePartitioning() {
	filepaths, err := s.repo.ListFilepathsByChannel(context.Background(), 999999)

	s.NoError(err)
	s.Empty(filepaths)
}

func TestVideoRepositorySuite(t *testing.T) {
	suite.Run(t, new(VideoRepositoryTestSuite))
}
