package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) SaveRefreshToken(ctx context.Context, userID int, token string) error {
	args := m.Called(ctx, userID, token)
	return args.Error(0)
}

func (m *MockTokenRepository) ValidateRefreshToken(ctx context.Context, token string) (int, error) {
	args := m.Called(ctx, token)
	return args.Int(0), args.Error(1)
}

func (m *MockTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepository) DeleteAllUserTokens(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockTokenRepository) BlacklistAccessToken(ctx context.Context, token string, ttl time.Duration) error {
	args := m.Called(ctx, token, ttl)
	return args.Error(0)
}

func (m *MockTokenRepository) IsAccessTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}
