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

type UserServiceTestSuite struct {
	suite.Suite
	mockRepo *mocks.UserRepository
	service  service.UserService
	mother   mother.UserMother
}

func (s *UserServiceTestSuite) SetupTest() {
	s.mockRepo = mocks.NewUserRepository(s.T())
	s.service = service.NewUserService(s.mockRepo)
	s.mother = mother.UserMother{}
}

func (s *UserServiceTestSuite) TestGetProfile_Positive_ReturnsUser() {
	ctx := context.Background()
	expectedUser := s.mother.ValidActiveUser()
	s.mockRepo.On("GetByID", ctx, expectedUser.ID).Return(expectedUser, nil)

	actualUser, err := s.service.GetProfile(ctx, expectedUser.ID)

	s.NoError(err)
	s.NotNil(actualUser)
	s.Equal(expectedUser.Username, actualUser.Username)
	s.mockRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestGetProfile_Negative_UserNotFoundException() {
	ctx := context.Background()
	s.mockRepo.On("GetByID", ctx, 999).Return(nil, nil)

	actualUser, err := s.service.GetProfile(ctx, 999)

	s.ErrorIs(err, domain.ErrUserNotFound)
	s.Nil(actualUser)
}

func (s *UserServiceTestSuite) TestGetProfile_Negative_UserIsBannedException() {
	ctx := context.Background()
	bannedUser := s.mother.BannedUser()
	s.mockRepo.On("GetByID", ctx, bannedUser.ID).Return(bannedUser, nil)

	actualUser, err := s.service.GetProfile(ctx, bannedUser.ID)

	s.ErrorIs(err, domain.ErrUserIsBanned)
	s.Nil(actualUser)
}

func (s *UserServiceTestSuite) TestSetNotificationsSettings_Positive_Success() {
	ctx := context.Background()
	userID := 1
	s.mockRepo.On("SetNotificationsEnabled", ctx, userID, false).Return(nil)

	err := s.service.SetNotificationsSettings(ctx, userID, false)

	s.NoError(err)
}

func (s *UserServiceTestSuite) TestSetNotificationsSettings_Negative_RepoError() {
	ctx := context.Background()
	expectedErr := errors.New("db connection lost")
	s.mockRepo.On("SetNotificationsEnabled", ctx, 1, true).Return(expectedErr)

	err := s.service.SetNotificationsSettings(ctx, 1, true)

	s.ErrorContains(err, "update notifications failed")
}

func (s *UserServiceTestSuite) TestDeleteAccount_Positive_Success() {
	ctx := context.Background()
	s.mockRepo.On("Delete", ctx, 1).Return(nil)

	err := s.service.DeleteAccount(ctx, 1)

	s.NoError(err)
}

func (s *UserServiceTestSuite) TestDeleteAccount_Negative_RepoError() {
	ctx := context.Background()
	s.mockRepo.On("Delete", ctx, 999).Return(domain.ErrUserNotFound)

	err := s.service.DeleteAccount(ctx, 999)

	s.ErrorIs(err, domain.ErrUserNotFound)
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
