package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockUserValidationService struct {
	mock.Mock
}

func (m *MockUserValidationService) ValidateNewUser(ctx context.Context, email, username, password string) error {
	args := m.Called(ctx, email, username, password)
	return args.Error(0)
}

func (m *MockUserValidationService) ValidatePassword(password string) error {
	args := m.Called(password)
	return args.Error(0)
}
