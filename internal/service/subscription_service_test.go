package service_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"ZVideo/internal/testing/mocks"
	"ZVideo/internal/testing/mother"
	"context"
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

func (s *SubscriptionServiceTestSuite) TestSubscribe_Positive_NewSubscription() {

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

func (s *SubscriptionServiceTestSuite) TestSubscribe_Negative_SelfSubscription() {
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

func TestSubscriptionServiceSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionServiceTestSuite))
}
