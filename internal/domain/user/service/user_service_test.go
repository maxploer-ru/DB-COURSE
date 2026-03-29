package service_test

import (
	"ZVideo/internal/domain/auth/entity"
	"ZVideo/internal/domain/user/repository/mocks"
	"ZVideo/internal/domain/user/service"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) ChangePassword(ctx context.Context, userID int, oldPassword, newPassword string) error {
	args := m.Called(ctx, userID, oldPassword, newPassword)
	return args.Error(0)
}

func (m *mockAuthService) RevokeAllUserTokens(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestUserService_GetProfile_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	svc := service.NewUserService(mockRepo)

	user := &entity.User{ID: 1, Username: "john", Email: "john@example.com"}
	mockRepo.On("GetByID", mock.Anything, 1).Return(user, nil)

	profile, err := svc.GetProfile(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "john", profile.Username)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateProfile_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	svc := service.NewUserService(mockRepo)

	existing := &entity.User{ID: 1, Username: "john", Email: "john@example.com"}
	mockRepo.On("GetByID", mock.Anything, 1).Return(existing, nil)
	mockRepo.On("ExistsByUsername", mock.Anything, "newuser").Return(false, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Username == "newuser"
	})).Return(nil)

	newUsername := "newuser"
	cmd := service.UpdateProfileCommand{Username: &newUsername}
	profile, err := svc.UpdateProfile(context.Background(), 1, cmd)
	assert.NoError(t, err)
	assert.Equal(t, "newuser", profile.Username)
}
