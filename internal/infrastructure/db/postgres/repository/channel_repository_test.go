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

type ChannelRepositoryTestSuite struct {
	suite.Suite
	pgContainer *db.PostgresContainer
	db          *gorm.DB
	tx          *gorm.DB
	repo        *repository.ChannelRepository
	userRepo    *repository.UserRepository
	userMother  mother.UserMother
	chanMother  mother.ChannelMother
	testUser    *domain.User
}

func (s *ChannelRepositoryTestSuite) SetupSuite() {
	s.db = sharedDB
}

func (s *ChannelRepositoryTestSuite) SetupTest() {
	s.tx = s.db.Begin()
	s.repo = repository.NewChannelRepository(s.tx)
	s.userRepo = repository.NewUserRepository(s.tx)
	s.userMother = mother.UserMother{}
	s.chanMother = mother.ChannelMother{}

	u := s.userMother.ValidActiveUser()
	u.ID = 0
	err := s.userRepo.Create(context.Background(), u)
	s.Require().NoError(err)
	s.testUser = u
}

func (s *ChannelRepositoryTestSuite) TearDownTest() {
	s.tx.Rollback()
}

func (s *ChannelRepositoryTestSuite) TestCreate_Positive() {
	ctx := context.Background()
	ch := s.chanMother.ChannelForUser(s.testUser.ID)
	ch.ID = 0

	err := s.repo.Create(ctx, ch)

	s.NoError(err)
	s.NotZero(ch.ID)
}

func (s *ChannelRepositoryTestSuite) TestCreate_Negative() {
	ctx := context.Background()
	ch1 := s.chanMother.ChannelForUser(s.testUser.ID)
	ch1.ID = 0
	_ = s.repo.Create(ctx, ch1)
	ch2 := s.chanMother.ChannelForUser(s.testUser.ID)
	ch2.ID = 0

	err := s.repo.Create(ctx, ch2)

	s.Error(err)
}

func (s *ChannelRepositoryTestSuite) TestGetByID_Positive() {
	ctx := context.Background()
	ch := s.chanMother.ChannelForUser(s.testUser.ID)
	ch.ID = 0
	_ = s.repo.Create(ctx, ch)

	found, err := s.repo.GetByID(ctx, ch.ID)

	s.NoError(err)
	s.NotNil(found)
	s.Equal(ch.Name, found.Name)
}

func (s *ChannelRepositoryTestSuite) TestGetByID_Negative() {
	ctx := context.Background()

	found, err := s.repo.GetByID(ctx, 999)

	s.NoError(err)
	s.Nil(found)
}

func (s *ChannelRepositoryTestSuite) TestGetByUserID_Positive() {
	ctx := context.Background()
	ch := s.chanMother.ChannelForUser(s.testUser.ID)
	ch.ID = 0
	_ = s.repo.Create(ctx, ch)

	found, err := s.repo.GetByUserID(ctx, s.testUser.ID)

	s.NoError(err)
	s.NotNil(found)
}

func (s *ChannelRepositoryTestSuite) TestGetByUserID_Negative() {
	ctx := context.Background()

	found, err := s.repo.GetByUserID(ctx, 999)

	s.NoError(err)
	s.Nil(found)
}

func (s *ChannelRepositoryTestSuite) TestGetByName_Positive() {
	ctx := context.Background()
	ch := s.chanMother.ChannelForUser(s.testUser.ID)
	ch.ID = 0
	_ = s.repo.Create(ctx, ch)

	found, err := s.repo.GetByName(ctx, ch.Name)

	s.NoError(err)
	s.NotNil(found)
}

func (s *ChannelRepositoryTestSuite) TestGetByName_Negative() {
	ctx := context.Background()

	found, err := s.repo.GetByName(ctx, "unknown")

	s.NoError(err)
	s.Nil(found)
}

func (s *ChannelRepositoryTestSuite) TestUpdate_Positive() {
	ctx := context.Background()
	ch := s.chanMother.ChannelForUser(s.testUser.ID)
	ch.ID = 0
	_ = s.repo.Create(ctx, ch)
	ch.Name = "updated_name"

	err := s.repo.Update(ctx, ch)

	s.NoError(err)
	found, _ := s.repo.GetByID(ctx, ch.ID)
	s.Equal("updated_name", found.Name)
}

func (s *ChannelRepositoryTestSuite) TestUpdate_Negative() {
	ctx := context.Background()
	ch := s.chanMother.ChannelForUser(s.testUser.ID)
	ch.ID = 0
	_ = s.repo.Create(ctx, ch)
	ch.UserID = 999999

	err := s.repo.Update(ctx, ch)

	s.Error(err)
}

func (s *ChannelRepositoryTestSuite) TestDelete_Positive() {
	ctx := context.Background()
	ch := s.chanMother.ChannelForUser(s.testUser.ID)
	ch.ID = 0
	_ = s.repo.Create(ctx, ch)

	err := s.repo.Delete(ctx, ch.ID)

	s.NoError(err)
	found, _ := s.repo.GetByID(ctx, ch.ID)
	s.Nil(found)
}

func (s *ChannelRepositoryTestSuite) TestDelete_Negative() {
	ctx := context.Background()

	err := s.repo.Delete(ctx, 999)

	s.NoError(err)
}

func (s *ChannelRepositoryTestSuite) TestExistsByName_Positive() {
	ctx := context.Background()
	ch := s.chanMother.ChannelForUser(s.testUser.ID)
	ch.ID = 0
	_ = s.repo.Create(ctx, ch)

	exists, err := s.repo.ExistsByName(ctx, ch.Name)

	s.NoError(err)
	s.True(exists)
}

func (s *ChannelRepositoryTestSuite) TestExistsByName_Negative() {
	ctx := context.Background()

	exists, err := s.repo.ExistsByName(ctx, "unknown")

	s.NoError(err)
	s.False(exists)
}

func (s *ChannelRepositoryTestSuite) TestListChannels_Positive() {
	ctx := context.Background()
	ch := s.chanMother.ChannelForUser(s.testUser.ID)
	ch.ID = 0
	_ = s.repo.Create(ctx, ch)

	list, err := s.repo.ListChannels(ctx, 10, 0)

	s.NoError(err)
	s.Len(list, 1)
}

func (s *ChannelRepositoryTestSuite) TestListChannels_Negative() {
	ctx := context.Background()

	list, err := s.repo.ListChannels(ctx, 10, 999)

	s.NoError(err)
	s.Len(list, 0)
}

func TestChannelRepositorySuite(t *testing.T) {
	suite.Run(t, new(ChannelRepositoryTestSuite))
}
