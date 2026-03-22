package mocks

import (
	"ZVideo/internal/domain/auth/service"

	"github.com/stretchr/testify/mock"
)

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateAccessToken(userID int, username, role string) (string, error) {
	args := m.Called(userID, username, role)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) GenerateRefreshToken(userID int) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateAccessToken(token string) (*service.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Claims), args.Error(1)
}

func (m *MockJWTService) ValidateRefreshToken(token string) (int, error) {
	args := m.Called(token)
	return args.Int(0), args.Error(1)
}
