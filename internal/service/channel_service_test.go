package service_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"ZVideo/internal/testing/mocks"
	"ZVideo/internal/testing/mother"
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ChannelServiceTestSuite struct {
	suite.Suite
	mockRepo *mocks.ChannelRepository
	service  service.ChannelService
	mother   mother.ChannelMother
}

func (s *ChannelServiceTestSuite) SetupTest() {
	s.mockRepo = mocks.NewChannelRepository(s.T())
	s.service = service.NewChannelService(s.mockRepo)
	s.mother = mother.ChannelMother{}
}

func (s *ChannelServiceTestSuite) TestCreateChannel_Positive() {
	ctx := context.Background()
	userID := 1
	name := "new_channel"
	desc := "desc"

	s.mockRepo.On("GetByUserID", ctx, userID).Return(nil, nil)
	s.mockRepo.On("ExistsByName", ctx, name).Return(false, nil)
	s.mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Channel")).Return(nil)

	ch, err := s.service.CreateChannel(ctx, userID, name, desc)

	s.NoError(err)
	s.NotNil(ch)
	s.Equal(name, ch.Name)
	s.mockRepo.AssertExpectations(s.T())
}

func (s *ChannelServiceTestSuite) TestCreateChannel_Negative() {
	ctx := context.Background()
	existingChannel := s.mother.ValidChannel()

	s.mockRepo.On("GetByUserID", ctx, existingChannel.UserID).Return(existingChannel, nil)

	ch, err := s.service.CreateChannel(ctx, existingChannel.UserID, "name", "desc")

	s.ErrorIs(err, domain.ErrChannelAlreadyExists)
	s.Nil(ch)
}

func (s *ChannelServiceTestSuite) TestGetChannel_Positive() {
	ctx := context.Background()
	expected := s.mother.ValidChannel()
	s.mockRepo.On("GetByID", ctx, expected.ID).Return(expected, nil)

	ch, err := s.service.GetChannel(ctx, expected.ID)

	s.NoError(err)
	s.Equal(expected.ID, ch.ID)
}

func (s *ChannelServiceTestSuite) TestGetChannel_Negative() {
	ctx := context.Background()
	s.mockRepo.On("GetByID", ctx, 999).Return(nil, nil)

	ch, err := s.service.GetChannel(ctx, 999)

	s.ErrorIs(err, domain.ErrChannelNotFound)
	s.Nil(ch)
}

func (s *ChannelServiceTestSuite) TestGetChannelByName_Positive() {
	ctx := context.Background()
	expected := s.mother.ValidChannel()
	s.mockRepo.On("GetByName", ctx, expected.Name).Return(expected, nil)

	ch, err := s.service.GetChannelByName(ctx, expected.Name)

	s.NoError(err)
	s.Equal(expected.Name, ch.Name)
}

func (s *ChannelServiceTestSuite) TestGetChannelByName_Negative() {
	ctx := context.Background()
	s.mockRepo.On("GetByName", ctx, "unknown").Return(nil, nil)

	ch, err := s.service.GetChannelByName(ctx, "unknown")

	s.ErrorIs(err, domain.ErrChannelNotFound)
	s.Nil(ch)
}

func (s *ChannelServiceTestSuite) TestGetChannelByUserID_Positive() {
	ctx := context.Background()
	expected := s.mother.ValidChannel()
	s.mockRepo.On("GetByUserID", ctx, expected.UserID).Return(expected, nil)

	ch, err := s.service.GetChannelByUserID(ctx, expected.UserID)

	s.NoError(err)
	s.Equal(expected.UserID, ch.UserID)
}

func (s *ChannelServiceTestSuite) TestGetChannelByUserID_Negative() {
	ctx := context.Background()
	s.mockRepo.On("GetByUserID", ctx, 999).Return(nil, nil)

	ch, err := s.service.GetChannelByUserID(ctx, 999)

	s.ErrorIs(err, domain.ErrChannelNotFound)
	s.Nil(ch)
}

func (s *ChannelServiceTestSuite) TestUpdateChannel_Positive() {
	ctx := context.Background()
	existing := s.mother.ValidChannel()
	newName := "updated_name"

	s.mockRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)
	s.mockRepo.On("ExistsByName", ctx, newName).Return(false, nil)
	s.mockRepo.On("Update", ctx, existing).Return(nil)

	ch, err := s.service.UpdateChannel(ctx, existing.ID, existing.UserID, &newName, nil)

	s.NoError(err)
	s.Equal(newName, ch.Name)
}

func (s *ChannelServiceTestSuite) TestUpdateChannel_Negative() {
	ctx := context.Background()
	existing := s.mother.ValidChannel()
	wrongUserID := 999
	newName := "updated_name"

	s.mockRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)

	ch, err := s.service.UpdateChannel(ctx, existing.ID, wrongUserID, &newName, nil)

	s.ErrorIs(err, domain.ErrForbidden)
	s.Nil(ch)
}

func (s *ChannelServiceTestSuite) TestDeleteChannel_Positive() {
	ctx := context.Background()
	existing := s.mother.ValidChannel()

	s.mockRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)
	s.mockRepo.On("Delete", ctx, existing.ID).Return(nil)

	err := s.service.DeleteChannel(ctx, existing.ID, existing.UserID)

	s.NoError(err)
}

func (s *ChannelServiceTestSuite) TestDeleteChannel_Negative() {
	ctx := context.Background()
	existing := s.mother.ValidChannel()
	wrongUserID := 999

	s.mockRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)

	err := s.service.DeleteChannel(ctx, existing.ID, wrongUserID)

	s.ErrorIs(err, domain.ErrForbidden)
}

func (s *ChannelServiceTestSuite) TestExists_Positive() {
	ctx := context.Background()
	existing := s.mother.ValidChannel()
	s.mockRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)

	exists, err := s.service.Exists(ctx, existing.ID)

	s.NoError(err)
	s.True(exists)
}

func (s *ChannelServiceTestSuite) TestExists_Negative() {
	ctx := context.Background()
	s.mockRepo.On("GetByID", ctx, 999).Return(nil, nil)

	exists, err := s.service.Exists(ctx, 999)

	s.NoError(err)
	s.False(exists)
}

func (s *ChannelServiceTestSuite) TestIsOwner_Positive() {
	ctx := context.Background()
	existing := s.mother.ValidChannel()
	s.mockRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)

	isOwner, err := s.service.IsOwner(ctx, existing.ID, existing.UserID)

	s.NoError(err)
	s.True(isOwner)
}

func (s *ChannelServiceTestSuite) TestIsOwner_Negative() {
	ctx := context.Background()
	existing := s.mother.ValidChannel()
	wrongUserID := 999
	s.mockRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)

	isOwner, err := s.service.IsOwner(ctx, existing.ID, wrongUserID)

	s.NoError(err)
	s.False(isOwner)
}

func (s *ChannelServiceTestSuite) TestListChannels_Positive() {
	ctx := context.Background()
	expected := []*domain.Channel{s.mother.ValidChannel()}
	s.mockRepo.On("ListChannels", ctx, 10, 0).Return(expected, nil)
	s.mockRepo.On("CountChannels", ctx).Return(int64(len(expected)), nil)

	channels, err := s.service.ListChannels(ctx, 10, 0)

	s.NoError(err)
	s.Len(channels.Items, 1)
	s.Equal(int64(1), channels.TotalCount)
}

func (s *ChannelServiceTestSuite) TestListChannels_Negative() {
	ctx := context.Background()
	s.mockRepo.On("ListChannels", ctx, 10, 0).Return(nil, domain.ErrInternalServer)

	channels, err := s.service.ListChannels(ctx, 10, 0)

	s.ErrorIs(err, domain.ErrInternalServer)
	s.Nil(channels)
}

func TestChannelServiceSuite(t *testing.T) {
	suite.Run(t, new(ChannelServiceTestSuite))
}
