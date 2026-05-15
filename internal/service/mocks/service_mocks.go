package mocks

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockPasswordService is a mock of service.PasswordService.
type MockPasswordService struct {
	mock.Mock
}

func (m *MockPasswordService) HashPassword(ctx context.Context, password string) (string, error) {
	args := m.Called(ctx, password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordService) ComparePassword(ctx context.Context, password, hash string) error {
	args := m.Called(ctx, password, hash)
	return args.Error(0)
}

// MockUserValidatorService is a mock of service.UserValidatorService.
type MockUserValidatorService struct {
	mock.Mock
}

func (m *MockUserValidatorService) ValidateNewUser(ctx context.Context, email, nickname, password string) error {
	args := m.Called(ctx, email, nickname, password)
	return args.Error(0)
}

// MockJWTService is a mock of service.JWTService.
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateAccessToken(ctx context.Context, data *domain.AccessTokenData) (string, error) {
	args := m.Called(ctx, data)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) GenerateRefreshToken(ctx context.Context, userID int) (string, *domain.RefreshTokenData, error) {
	args := m.Called(ctx, userID)
	if data, ok := args.Get(1).(*domain.RefreshTokenData); ok {
		return args.String(0), data, args.Error(2)
	}
	return args.String(0), nil, args.Error(2)
}

func (m *MockJWTService) ValidateAccessToken(ctx context.Context, accessToken string) (*domain.AccessTokenData, error) {
	args := m.Called(ctx, accessToken)
	if data, ok := args.Get(0).(*domain.AccessTokenData); ok {
		return data, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockJWTService) ValidateRefreshToken(ctx context.Context, refreshToken string) (*domain.RefreshTokenData, error) {
	args := m.Called(ctx, refreshToken)
	if data, ok := args.Get(0).(*domain.RefreshTokenData); ok {
		return data, args.Error(1)
	}
	return nil, args.Error(1)
}

// MockChannelService is a mock of service.ChannelService.
type MockChannelService struct {
	mock.Mock
}

func (m *MockChannelService) CreateChannel(ctx context.Context, userID int, name, description string) (*domain.Channel, error) {
	args := m.Called(ctx, userID, name, description)
	if channel, ok := args.Get(0).(*domain.Channel); ok {
		return channel, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockChannelService) GetChannel(ctx context.Context, id int) (*domain.Channel, error) {
	args := m.Called(ctx, id)
	if channel, ok := args.Get(0).(*domain.Channel); ok {
		return channel, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockChannelService) GetChannelByName(ctx context.Context, name string) (*domain.Channel, error) {
	args := m.Called(ctx, name)
	if channel, ok := args.Get(0).(*domain.Channel); ok {
		return channel, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockChannelService) GetChannelByUserID(ctx context.Context, userID int) (*domain.Channel, error) {
	args := m.Called(ctx, userID)
	if channel, ok := args.Get(0).(*domain.Channel); ok {
		return channel, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockChannelService) UpdateChannel(ctx context.Context, channelID, userID int, name, description *string) (*domain.Channel, error) {
	args := m.Called(ctx, channelID, userID, name, description)
	if channel, ok := args.Get(0).(*domain.Channel); ok {
		return channel, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockChannelService) DeleteChannel(ctx context.Context, channelID, userID int) error {
	args := m.Called(ctx, channelID, userID)
	return args.Error(0)
}

func (m *MockChannelService) Exists(ctx context.Context, channelID int) (bool, error) {
	args := m.Called(ctx, channelID)
	return args.Bool(0), args.Error(1)
}

func (m *MockChannelService) IsOwner(ctx context.Context, channelID, userID int) (bool, error) {
	args := m.Called(ctx, channelID, userID)
	return args.Bool(0), args.Error(1)
}

// MockStorageService is a mock of service.StorageService.
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) GenerateUploadPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, key, expiry)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) GenerateAccessPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, key, expiry)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) DeleteObject(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

var _ service.PasswordService = (*MockPasswordService)(nil)
var _ service.UserValidatorService = (*MockUserValidatorService)(nil)
var _ service.JWTService = (*MockJWTService)(nil)
var _ service.ChannelService = (*MockChannelService)(nil)
var _ service.StorageService = (*MockStorageService)(nil)
