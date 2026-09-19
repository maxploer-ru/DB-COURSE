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

type PlaylistRepositoryTestSuite struct {
	suite.Suite
	pgContainer *db.PostgresContainer
	db          *gorm.DB
	tx          *gorm.DB
	repo        *repository.PlaylistRepository
	userRepo    *repository.UserRepository
	chanRepo    *repository.ChannelRepository
	videoRepo   *repository.VideoRepository
	userMother  mother.UserMother
	chanMother  mother.ChannelMother
	vidMother   mother.VideoMother
	plMother    mother.PlaylistMother
	testUser    *domain.User
	testChannel *domain.Channel
	testVideo   *domain.Video
}

func (s *PlaylistRepositoryTestSuite) SetupSuite() {
	s.db = sharedDB
}

func (s *PlaylistRepositoryTestSuite) SetupTest() {
	s.tx = s.db.Begin()
	s.repo = repository.NewPlaylistRepository(s.tx)
	s.userRepo = repository.NewUserRepository(s.tx)
	s.chanRepo = repository.NewChannelRepository(s.tx)
	s.videoRepo = repository.NewVideoRepository(s.tx)

	s.userMother = mother.UserMother{}
	s.chanMother = mother.ChannelMother{}
	s.vidMother = mother.VideoMother{}
	s.plMother = mother.PlaylistMother{}

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

	v := s.vidMother.ReadyVideoForChannel(s.testChannel.ID)
	v.ID = 0
	err = s.videoRepo.Create(context.Background(), v)
	s.Require().NoError(err)
	s.testVideo = v
}

func (s *PlaylistRepositoryTestSuite) TearDownTest() {
	s.tx.Rollback()
}

func (s *PlaylistRepositoryTestSuite) TestCreate_Positive_StateTransition() {
	ctx := context.Background()
	pl := s.plMother.PlaylistForChannel(s.testChannel.ID)
	pl.ID = 0

	err := s.repo.Create(ctx, pl)

	s.NoError(err)
	s.NotZero(pl.ID)
}

func (s *PlaylistRepositoryTestSuite) TestCreate_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	pl := s.plMother.PlaylistForChannel(999999)
	pl.ID = 0

	err := s.repo.Create(ctx, pl)

	s.Error(err)
}

func (s *PlaylistRepositoryTestSuite) TestGetByID_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	pl := s.plMother.PlaylistForChannel(s.testChannel.ID)
	pl.ID = 0
	_ = s.repo.Create(ctx, pl)

	found, err := s.repo.GetByID(ctx, pl.ID)

	s.NoError(err)
	s.NotNil(found)
	s.Equal(pl.Name, found.Name)
}

func (s *PlaylistRepositoryTestSuite) TestGetByID_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	found, err := s.repo.GetByID(ctx, 999)

	s.NoError(err)
	s.Nil(found)
}

func (s *PlaylistRepositoryTestSuite) TestListByChannel_Positive_BoundaryValueAnalysis() {
	ctx := context.Background()
	pl := s.plMother.PlaylistForChannel(s.testChannel.ID)
	pl.ID = 0
	_ = s.repo.Create(ctx, pl)

	list, err := s.repo.ListByChannel(ctx, s.testChannel.ID, 10, 0)

	s.NoError(err)
	s.Len(list, 1)
}

func (s *PlaylistRepositoryTestSuite) TestListByChannel_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	list, err := s.repo.ListByChannel(ctx, 999, 10, 0)

	s.NoError(err)
	s.Len(list, 0)
}

func (s *PlaylistRepositoryTestSuite) TestUpdate_Positive_StateTransition() {
	ctx := context.Background()
	pl := s.plMother.PlaylistForChannel(s.testChannel.ID)
	pl.ID = 0
	_ = s.repo.Create(ctx, pl)
	pl.Name = "Updated"

	err := s.repo.Update(ctx, pl)

	s.NoError(err)
	found, _ := s.repo.GetByID(ctx, pl.ID)
	s.Equal("Updated", found.Name)
}

func (s *PlaylistRepositoryTestSuite) TestUpdate_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	pl := s.plMother.PlaylistForChannel(s.testChannel.ID)
	pl.ID = 999999

	err := s.repo.Update(ctx, pl)

	s.ErrorIs(err, domain.ErrPlaylistNotFound)
}

func (s *PlaylistRepositoryTestSuite) TestDelete_Positive_StateTransition() {
	ctx := context.Background()
	pl := s.plMother.PlaylistForChannel(s.testChannel.ID)
	pl.ID = 0
	_ = s.repo.Create(ctx, pl)

	err := s.repo.Delete(ctx, pl.ID)

	s.NoError(err)
	found, _ := s.repo.GetByID(ctx, pl.ID)
	s.Nil(found)
}

func (s *PlaylistRepositoryTestSuite) TestDelete_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	err := s.repo.Delete(ctx, 999)

	s.ErrorIs(err, domain.ErrPlaylistNotFound)
}

func (s *PlaylistRepositoryTestSuite) TestAddVideo_Positive_StateTransition() {
	ctx := context.Background()
	pl := s.plMother.PlaylistForChannel(s.testChannel.ID)
	pl.ID = 0
	_ = s.repo.Create(ctx, pl)

	err := s.repo.AddVideo(ctx, pl.ID, s.testVideo.ID)

	s.NoError(err)
	count, _ := s.repo.GetItemsCount(ctx, pl.ID)
	s.Equal(1, count)
}

func (s *PlaylistRepositoryTestSuite) TestAddVideo_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	pl := s.plMother.PlaylistForChannel(s.testChannel.ID)
	pl.ID = 0
	_ = s.repo.Create(ctx, pl)

	err := s.repo.AddVideo(ctx, pl.ID, 999999)

	s.Error(err)
}

func (s *PlaylistRepositoryTestSuite) TestRemoveVideo_Positive_StateTransition() {
	ctx := context.Background()
	pl := s.plMother.PlaylistForChannel(s.testChannel.ID)
	pl.ID = 0
	_ = s.repo.Create(ctx, pl)
	_ = s.repo.AddVideo(ctx, pl.ID, s.testVideo.ID)

	err := s.repo.RemoveVideo(ctx, pl.ID, s.testVideo.ID)

	s.NoError(err)
	count, _ := s.repo.GetItemsCount(ctx, pl.ID)
	s.Equal(0, count)
}

func (s *PlaylistRepositoryTestSuite) TestRemoveVideo_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	err := s.repo.RemoveVideo(ctx, 999, 999)

	s.NoError(err)
}

func (s *PlaylistRepositoryTestSuite) TestUpdateVideoPosition_Positive_StateTransition() {
	ctx := context.Background()
	pl := s.plMother.PlaylistForChannel(s.testChannel.ID)
	pl.ID = 0
	_ = s.repo.Create(ctx, pl)
	_ = s.repo.AddVideo(ctx, pl.ID, s.testVideo.ID)

	v2 := s.vidMother.ReadyVideoForChannel(s.testChannel.ID)
	v2.ID = 0
	_ = s.videoRepo.Create(ctx, v2)
	_ = s.repo.AddVideo(ctx, pl.ID, v2.ID)

	err := s.repo.UpdateVideoPosition(ctx, pl.ID, v2.ID, 1)

	s.NoError(err)
	items, _ := s.repo.ListItems(ctx, pl.ID, 10, 0)
	s.Equal(v2.ID, items[0].VideoID)
	s.Equal(1, items[0].Number)
}

func (s *PlaylistRepositoryTestSuite) TestUpdateVideoPosition_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	err := s.repo.UpdateVideoPosition(ctx, 999, 999, 1)

	s.Error(err)
}

func (s *PlaylistRepositoryTestSuite) TestListItems_Positive_BoundaryValueAnalysis() {
	ctx := context.Background()
	pl := s.plMother.PlaylistForChannel(s.testChannel.ID)
	pl.ID = 0
	_ = s.repo.Create(ctx, pl)
	_ = s.repo.AddVideo(ctx, pl.ID, s.testVideo.ID)

	items, err := s.repo.ListItems(ctx, pl.ID, 10, 0)

	s.NoError(err)
	s.Len(items, 1)
	s.Equal(s.testVideo.Title, items[0].VideoTitle)
}

func (s *PlaylistRepositoryTestSuite) TestListItems_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	items, err := s.repo.ListItems(ctx, 999, 10, 0)

	s.NoError(err)
	s.Len(items, 0)
}

func (s *PlaylistRepositoryTestSuite) TestGetItemsCount_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	pl := s.plMother.PlaylistForChannel(s.testChannel.ID)
	pl.ID = 0
	_ = s.repo.Create(ctx, pl)
	_ = s.repo.AddVideo(ctx, pl.ID, s.testVideo.ID)

	count, err := s.repo.GetItemsCount(ctx, pl.ID)

	s.NoError(err)
	s.Equal(1, count)
}

func (s *PlaylistRepositoryTestSuite) TestGetItemsCount_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	count, err := s.repo.GetItemsCount(ctx, 999)

	s.NoError(err)
	s.Equal(0, count)
}

func TestPlaylistRepositorySuite(t *testing.T) {
	suite.Run(t, new(PlaylistRepositoryTestSuite))
}
