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
	pgContainer    *db.PostgresContainer
	db             *gorm.DB
	tx             *gorm.DB
	repo           *repository.SubscriptionRepository
	subMother      mother.SubscriptionMother
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

func (s *SubscriptionRepositoryTestSuite) TestSubscribe_Negative() {
	ctx := context.Background()

	created, err := s.repo.Subscribe(ctx, s.testSubscriber.ID, 999999)

	s.Error(err)
	s.False(created)
}

func (s *SubscriptionRepositoryTestSuite) TestUnsubscribe_Positive() {
	ctx := context.Background()
	_, _ = s.repo.Subscribe(ctx, s.testSubscriber.ID, s.testChannel.ID)

	deleted, err := s.repo.Unsubscribe(ctx, s.testSubscriber.ID, s.testChannel.ID)

	s.NoError(err)
	s.True(deleted)
}

func (s *SubscriptionRepositoryTestSuite) TestUnsubscribe_Negative() {
	ctx := context.Background()

	deleted, err := s.repo.Unsubscribe(ctx, s.testSubscriber.ID, 999999)

	s.NoError(err)
	s.False(deleted)
}

func (s *SubscriptionRepositoryTestSuite) TestIsSubscribed_Positive() {
	ctx := context.Background()
	_, _ = s.repo.Subscribe(ctx, s.testSubscriber.ID, s.testChannel.ID)

	ok, err := s.repo.IsSubscribed(ctx, s.testSubscriber.ID, s.testChannel.ID)

	s.NoError(err)
	s.True(ok)
}

func (s *SubscriptionRepositoryTestSuite) TestIsSubscribed_Negative() {
	ctx := context.Background()

	ok, err := s.repo.IsSubscribed(ctx, s.testSubscriber.ID, 999999)

	s.NoError(err)
	s.False(ok)
}

func (s *SubscriptionRepositoryTestSuite) TestGetSubscribersCount_Positive() {
	ctx := context.Background()
	_, _ = s.repo.Subscribe(ctx, s.testSubscriber.ID, s.testChannel.ID)

	count, err := s.repo.GetSubscribersCount(ctx, s.testChannel.ID)

	s.NoError(err)
	s.Equal(1, count)
}

func (s *SubscriptionRepositoryTestSuite) TestGetSubscribersCount_Negative() {
	ctx := context.Background()

	count, err := s.repo.GetSubscribersCount(ctx, 999999)

	s.NoError(err)
	s.Equal(0, count)
}

func (s *SubscriptionRepositoryTestSuite) TestGetUserSubscriptions_Positive() {
	ctx := context.Background()
	_, _ = s.repo.Subscribe(ctx, s.testSubscriber.ID, s.testChannel.ID)

	subs, err := s.repo.GetUserSubscriptions(ctx, s.testSubscriber.ID, 10, 0)

	s.NoError(err)
	s.Len(subs, 1)
	s.Equal(s.testChannel.Name, subs[0].ChannelName)
}

func (s *SubscriptionRepositoryTestSuite) TestGetUserSubscriptions_Negative() {
	ctx := context.Background()

	subs, err := s.repo.GetUserSubscriptions(ctx, 999999, 10, 0)

	s.NoError(err)
	s.Len(subs, 0)
}

func (s *SubscriptionRepositoryTestSuite) TestNotifySubscribersAboutNewVideo_Positive() {
	ctx := context.Background()
	_, _ = s.repo.Subscribe(ctx, s.testSubscriber.ID, s.testChannel.ID)

	err := s.repo.NotifySubscribersAboutNewVideo(ctx, s.testChannel.ID)

	s.NoError(err)
	subs, _ := s.repo.GetUserSubscriptions(ctx, s.testSubscriber.ID, 10, 0)
	s.Equal(1, subs[0].NewVideosCount)
}

func (s *SubscriptionRepositoryTestSuite) TestNotifySubscribersAboutNewVideo_Negative() {
	ctx := context.Background()

	err := s.repo.NotifySubscribersAboutNewVideo(ctx, 999999)

	s.NoError(err)
}

func (s *SubscriptionRepositoryTestSuite) TestResetNewVideosCount_Positive() {
	ctx := context.Background()
	_, _ = s.repo.Subscribe(ctx, s.testSubscriber.ID, s.testChannel.ID)
	_ = s.repo.NotifySubscribersAboutNewVideo(ctx, s.testChannel.ID)

	err := s.repo.ResetNewVideosCount(ctx, s.testSubscriber.ID, s.testChannel.ID)

	s.NoError(err)
	subs, _ := s.repo.GetUserSubscriptions(ctx, s.testSubscriber.ID, 10, 0)
	s.Equal(0, subs[0].NewVideosCount)
}

func (s *SubscriptionRepositoryTestSuite) TestResetNewVideosCount_Negative() {
	ctx := context.Background()

	err := s.repo.ResetNewVideosCount(ctx, 999999, s.testChannel.ID)

	s.NoError(err)
}

func TestSubscriptionRepositorySuite(t *testing.T) {
	suite.Run(t, new(SubscriptionRepositoryTestSuite))
}
