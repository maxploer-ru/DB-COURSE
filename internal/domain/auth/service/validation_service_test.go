package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ZVideo/internal/domain/auth/repository/mocks"
	"ZVideo/internal/domain/auth/service"
	"ZVideo/internal/infrastructure/config"
)

func TestUserValidationService_ValidateNewUser_Success(t *testing.T) {

	mockUserRepo := new(mocks.MockUserRepository)

	cfg := config.PasswordConfig{
		MinLength: 8,
	}

	mockUserRepo.On("ExistsByEmail", mock.Anything, "john@example.com").
		Return(false, nil)

	mockUserRepo.On("ExistsByUsername", mock.Anything, "john").
		Return(false, nil)

	valSvc := service.NewUserValidationService(mockUserRepo, cfg)

	err := valSvc.ValidateNewUser(
		context.Background(),
		"john@example.com",
		"john",
		"StrongPass123",
	)

	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestUserValidationService_ValidateNewUser_WeakPassword_TooShort(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)

	cfg := config.PasswordConfig{
		MinLength: 8,
	}

	valSvc := service.NewUserValidationService(mockUserRepo, cfg)

	err := valSvc.ValidateNewUser(
		context.Background(),
		"john@example.com",
		"john",
		"Short1",
	)

	assert.Error(t, err)
	assert.Equal(t, service.ErrWeakPassword, err)

	mockUserRepo.AssertNotCalled(t, "ExistsByEmail", mock.Anything, mock.Anything)
}

func TestUserValidationService_ValidateNewUser_EmailAlreadyExists(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)

	cfg := config.PasswordConfig{
		MinLength: 8,
	}

	mockUserRepo.On("ExistsByEmail", mock.Anything, "existing@example.com").
		Return(true, nil)

	valSvc := service.NewUserValidationService(mockUserRepo, cfg)

	err := valSvc.ValidateNewUser(
		context.Background(),
		"existing@example.com",
		"john",
		"StrongPass123",
	)

	assert.Error(t, err)
	assert.Equal(t, service.ErrEmailAlreadyExists, err)

	mockUserRepo.AssertNotCalled(t, "ExistsByUsername", mock.Anything, mock.Anything)
}

func TestUserValidationService_ValidateNewUser_UsernameAlreadyExists(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)

	cfg := config.PasswordConfig{
		MinLength: 8,
	}

	mockUserRepo.On("ExistsByEmail", mock.Anything, "john@example.com").
		Return(false, nil)

	mockUserRepo.On("ExistsByUsername", mock.Anything, "john").
		Return(true, nil)

	valSvc := service.NewUserValidationService(mockUserRepo, cfg)

	err := valSvc.ValidateNewUser(
		context.Background(),
		"john@example.com",
		"john",
		"StrongPass123",
	)

	assert.Error(t, err)
	assert.Equal(t, service.ErrUsernameAlreadyExists, err)
}
