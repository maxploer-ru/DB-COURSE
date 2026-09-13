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

type SubscriptionRepositoryTestSuite struct {
	suite.Suite
	pgContainer *db.PostgresContainer
	db          *gorm.DB
	tx          *gorm.DB
	repo        *repository.SubscriptionRepository
	subMother   mother.SubscriptionMother

	testSubscriber *models.User
	testChannel    *models.Channel
}

func (s *SubscriptionRepositoryTestSuite) SetupSuite() {
	s.db = sharedDB
	s.subMother = mother.SubscriptionMother{}
}

func (s *SubscriptionRepositoryTestSuite) SetupTest() {
	s.tx = s.db.Begin()
	s.repo = repository.NewSubscriptionRepository(s.tx)

	s.testSubscriber = &models.User{Username: "sub_user", Email: "1@t.com", PasswordHash: "x", RoleID: 1}
	owner := &models.User{Username: "owner", Email: "2@t.com", PasswordHash: "x", RoleID: 1}
	s.tx.Create(s.testSubscriber)
	s.tx.Create(owner)

	s.testChannel = &models.Channel{UserID: owner.ID, Name: "Target Channel"}
	s.tx.Create(s.testChannel)
}

func (s *SubscriptionRepositoryTestSuite) TearDownTest() {
	s.tx.Rollback()
}

func (s *SubscriptionRepositoryTestSuite) TestSubscribe_Positive() {
	ctx := context.Background()

	created, err := s.repo.Subscribe(ctx, s.testSubscriber.ID, s.testChannel.ID)

	s.NoError(err)
	s.True(created)

	createdAgain, err := s.repo.Subscribe(ctx, s.testSubscriber.ID, s.testChannel.ID)
	s.NoError(err)
	s.False(createdAgain)
}

func (s *SubscriptionRepositoryTestSuite) TestGetUserSubscriptions_Positive_ReturnsWithChannelName() {
	ctx := context.Background()
	_, _ = s.repo.Subscribe(ctx, s.testSubscriber.ID, s.testChannel.ID)

	subs, err := s.repo.GetUserSubscriptions(ctx, s.testSubscriber.ID, 10, 0)

	s.NoError(err)
	s.Len(subs, 1)
	s.Equal(s.testChannel.Name, subs[0].ChannelName)
}

func TestSubscriptionRepositorySuite(t *testing.T) {
	suite.Run(t, new(SubscriptionRepositoryTestSuite))
}
