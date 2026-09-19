package service_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"ZVideo/internal/testing/mocks"
	"ZVideo/internal/testing/mother"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SubscriptionServiceTestSuite struct {
	suite.Suite
	mockSubRepo     *mocks.SubscriptionRepository
	mockChannelRepo *mocks.ChannelRepository
	mockCounter     *mocks.SubscriberCounter
	service         service.SubscriptionService
	mother          mother.SubscriptionMother
	chanMother      mother.ChannelMother
}

func (s *SubscriptionServiceTestSuite) SetupTest() {
	s.mockSubRepo = mocks.NewSubscriptionRepository(s.T())
	s.mockChannelRepo = mocks.NewChannelRepository(s.T())
	s.mockCounter = mocks.NewSubscriberCounter(s.T())
	s.service = service.NewSubscriptionService(s.mockSubRepo, s.mockChannelRepo, s.mockCounter)
	s.mother = mother.SubscriptionMother{}
	s.chanMother = mother.ChannelMother{}
}

func (s *SubscriptionServiceTestSuite) TestSubscribe_Positive_NewSubscription_StateTransition() {
	ctx := context.Background()
	sub := s.mother.ValidSubscription()
	channel := s.chanMother.ValidChannel()
	channel.ID = sub.ChannelID
	channel.UserID = 999

	s.mockChannelRepo.On("GetByID", ctx, sub.ChannelID).Return(channel, nil)
	s.mockSubRepo.On("Subscribe", ctx, sub.UserID, sub.ChannelID).Return(true, nil)
	s.mockCounter.On("Increment", ctx, sub.ChannelID).Return(nil)

	err := s.service.Subscribe(ctx, sub.UserID, sub.ChannelID)

	s.NoError(err)
	s.mockSubRepo.AssertExpectations(s.T())
	s.mockCounter.AssertExpectations(s.T())
}

func (s *SubscriptionServiceTestSuite) TestSubscribe_Negative_SelfSubscription_Combinatorial() {
	ctx := context.Background()
	sub := s.mother.SelfSubscription()
	channel := s.chanMother.ValidChannel()
	channel.ID = sub.ChannelID
	channel.UserID = sub.UserID

	s.mockChannelRepo.On("GetByID", ctx, sub.ChannelID).Return(channel, nil)

	err := s.service.Subscribe(ctx, sub.UserID, sub.ChannelID)

	s.ErrorIs(err, domain.ErrSelfSubscription)
	s.mockSubRepo.AssertNotCalled(s.T(), "Subscribe")
}

func (s *SubscriptionServiceTestSuite) TestSubscribe_Negative_ChannelNotFound_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockChannelRepo.On("GetByID", ctx, 999).Return(nil, nil)

	err := s.service.Subscribe(ctx, 1, 999)

	s.ErrorIs(err, domain.ErrChannelNotFound)
}

func (s *SubscriptionServiceTestSuite) TestUnsubscribe_Positive_StateTransition() {
	ctx := context.Background()

	s.mockSubRepo.On("Unsubscribe", ctx, 1, 2).Return(true, nil)
	s.mockCounter.On("Decrement", ctx, 2).Return(nil)

	err := s.service.Unsubscribe(ctx, 1, 2)

	s.NoError(err)
}

func (s *SubscriptionServiceTestSuite) TestUnsubscribe_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockSubRepo.On("Unsubscribe", ctx, 1, 2).Return(false, errors.New("db error"))

	err := s.service.Unsubscribe(ctx, 1, 2)

	s.Error(err)
}

func (s *SubscriptionServiceTestSuite) TestIsSubscribed_Positive_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockSubRepo.On("IsSubscribed", ctx, 1, 2).Return(true, nil)

	ok, err := s.service.IsSubscribed(ctx, 1, 2)

	s.NoError(err)
	s.True(ok)
}

func (s *SubscriptionServiceTestSuite) TestIsSubscribed_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockSubRepo.On("IsSubscribed", ctx, 1, 2).Return(false, errors.New("db error"))

	ok, err := s.service.IsSubscribed(ctx, 1, 2)

	s.Error(err)
	s.False(ok)
}

func (s *SubscriptionServiceTestSuite) TestGetSubscribersCount_Positive_CacheHit_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockCounter.On("Get", ctx, 1).Return(42, true, nil)

	count, err := s.service.GetSubscribersCount(ctx, 1)

	s.NoError(err)
	s.Equal(42, count)
}

func (s *SubscriptionServiceTestSuite) TestGetSubscribersCount_Positive_DBFallback_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockCounter.On("Get", ctx, 1).Return(0, false, nil)
	s.mockSubRepo.On("GetSubscribersCount", ctx, 1).Return(15, nil)
	s.mockCounter.On("Set", ctx, 1, 15).Return(nil)

	count, err := s.service.GetSubscribersCount(ctx, 1)

	s.NoError(err)
	s.Equal(15, count)
}

func (s *SubscriptionServiceTestSuite) TestGetSubscribersCount_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockCounter.On("Get", ctx, 1).Return(0, false, nil)
	s.mockSubRepo.On("GetSubscribersCount", ctx, 1).Return(0, errors.New("db error"))

	count, err := s.service.GetSubscribersCount(ctx, 1)

	s.Error(err)
	s.Equal(0, count)
}

func (s *SubscriptionServiceTestSuite) TestGetUserSubscriptions_Positive_BoundaryValueAnalysis() {
	ctx := context.Background()
	expected := []*domain.Subscription{s.mother.ValidSubscription()}

	s.mockSubRepo.On("GetUserSubscriptions", ctx, 1, 10, 0).Return(expected, nil)
	s.mockSubRepo.On("CountUserSubscriptions", ctx, 1).Return(int64(len(expected)), nil)

	subs, err := s.service.GetUserSubscriptions(ctx, 1, 10, 0)

	s.NoError(err)
	s.Len(subs.Items, 1)
	s.Equal(int64(1), subs.TotalCount)
}

func (s *SubscriptionServiceTestSuite) TestGetUserSubscriptions_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockSubRepo.On("GetUserSubscriptions", ctx, 1, 10, 0).Return(nil, errors.New("db error"))

	subs, err := s.service.GetUserSubscriptions(ctx, 1, 10, 0)

	s.Error(err)
	s.Nil(subs)
}

func (s *SubscriptionServiceTestSuite) TestResetNewVideosCount_Positive_StateTransition() {
	ctx := context.Background()

	s.mockSubRepo.On("ResetNewVideosCount", ctx, 1, 2).Return(nil)

	err := s.service.ResetNewVideosCount(ctx, 1, 2)

	s.NoError(err)
}

func (s *SubscriptionServiceTestSuite) TestResetNewVideosCount_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockSubRepo.On("ResetNewVideosCount", ctx, 1, 2).Return(errors.New("db error"))

	err := s.service.ResetNewVideosCount(ctx, 1, 2)

	s.Error(err)
}

func (s *SubscriptionServiceTestSuite) TestNotifyAboutNewVideo_Positive_StateTransition() {
	ctx := context.Background()

	s.mockSubRepo.On("NotifySubscribersAboutNewVideo", ctx, 1).Return(nil)

	err := s.service.NotifyAboutNewVideo(ctx, 1)

	s.NoError(err)
}

func (s *SubscriptionServiceTestSuite) TestNotifyAboutNewVideo_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockSubRepo.On("NotifySubscribersAboutNewVideo", ctx, 1).Return(errors.New("db error"))

	err := s.service.NotifyAboutNewVideo(ctx, 1)

	s.Error(err)
}

func TestSubscriptionServiceSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionServiceTestSuite))
}
