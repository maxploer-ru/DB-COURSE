package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"ZVideo/internal/domain/auth/entity"
	"ZVideo/internal/domain/auth/repository/mocks"
	authSvc "ZVideo/internal/domain/auth/service"
	svcMocks "ZVideo/internal/domain/auth/service/mocks"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_Login_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockTokenRepo := new(mocks.MockTokenRepository)
	mockPasswordSvc := new(svcMocks.MockPasswordService)
	mockJwtSvc := new(svcMocks.MockJWTService)
	mockUserValSvc := new(svcMocks.MockUserValidationService)

	svc := authSvc.NewAuthService(mockUserRepo, mockTokenRepo, mockPasswordSvc, mockJwtSvc, mockUserValSvc)

	user := &entity.User{
		ID:           1,
		Username:     "john",
		Email:        "john@example.com",
		PasswordHash: "hashed",
		IsActive:     true,
		Role:         &entity.Role{Name: "user"},
	}
	mockUserRepo.On("GetByEmail", mock.Anything, "john@example.com").Return(user, nil)
	mockPasswordSvc.On("Verify", "hashed", "pass").Return(true)
	mockJwtSvc.On("GenerateAccessToken", user.ID, user.Username, user.Role.Name).Return("access", nil)
	mockJwtSvc.On("GenerateRefreshToken", user.ID).Return("refresh", nil)
	mockTokenRepo.On("SaveRefreshToken", mock.Anything, user.ID, "refresh").Return(nil)

	result, err := svc.Login(context.Background(), "john@example.com", "pass")
	assert.NoError(t, err)
	assert.Equal(t, "access", result.AccessToken)
	assert.Equal(t, "refresh", result.RefreshToken)
	assert.Equal(t, user, result.User)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
	mockPasswordSvc.AssertExpectations(t)
	mockJwtSvc.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockTokenRepo := new(mocks.MockTokenRepository)
	mockPasswordSvc := new(svcMocks.MockPasswordService)
	mockJwtSvc := new(svcMocks.MockJWTService)
	mockUserValSvc := new(svcMocks.MockUserValidationService)

	svc := authSvc.NewAuthService(mockUserRepo, mockTokenRepo, mockPasswordSvc, mockJwtSvc, mockUserValSvc)

	mockUserRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, nil)
	_, err := svc.Login(context.Background(), "notfound@example.com", "pass")
	assert.ErrorIs(t, err, authSvc.ErrInvalidCredentials)

	user := &entity.User{PasswordHash: "hashed"}
	mockUserRepo.On("GetByEmail", mock.Anything, "wrong@example.com").Return(user, nil)
	mockPasswordSvc.On("Verify", "hashed", "bad").Return(false)
	_, err = svc.Login(context.Background(), "wrong@example.com", "bad")
	assert.ErrorIs(t, err, authSvc.ErrInvalidCredentials)

	inactive := &entity.User{IsActive: false}
	mockUserRepo.On("GetByEmail", mock.Anything, "inactive@example.com").Return(inactive, nil)
	mockPasswordSvc.On("Verify", "", "pass").Return(true)
	_, err = svc.Login(context.Background(), "inactive@example.com", "pass")
	assert.ErrorIs(t, err, authSvc.ErrUserInactive)
}

func TestAuthService_Refresh_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockTokenRepo := new(mocks.MockTokenRepository)
	mockPasswordSvc := new(svcMocks.MockPasswordService)
	mockJwtSvc := new(svcMocks.MockJWTService)
	mockUserValSvc := new(svcMocks.MockUserValidationService)

	svc := authSvc.NewAuthService(mockUserRepo, mockTokenRepo, mockPasswordSvc, mockJwtSvc, mockUserValSvc)

	mockTokenRepo.On("ValidateRefreshToken", mock.Anything, "oldRefresh").Return(1, nil)
	user := &entity.User{ID: 1, Username: "john", Role: &entity.Role{Name: "user"}, IsActive: true}
	mockUserRepo.On("GetByID", mock.Anything, 1).Return(user, nil)
	mockJwtSvc.On("GenerateAccessToken", user.ID, user.Username, user.Role.Name).Return("newAccess", nil)
	mockJwtSvc.On("GenerateRefreshToken", user.ID).Return("newRefresh", nil)
	mockTokenRepo.On("DeleteRefreshToken", mock.Anything, "oldRefresh").Return(nil)
	mockTokenRepo.On("SaveRefreshToken", mock.Anything, user.ID, "newRefresh").Return(nil)

	result, err := svc.Refresh(context.Background(), "oldRefresh")
	assert.NoError(t, err)
	assert.Equal(t, "newAccess", result.AccessToken)
	assert.Equal(t, "newRefresh", result.RefreshToken)
}

func TestAuthService_Refresh_InvalidToken(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockTokenRepo := new(mocks.MockTokenRepository)
	mockPasswordSvc := new(svcMocks.MockPasswordService)
	mockJwtSvc := new(svcMocks.MockJWTService)
	mockUserValSvc := new(svcMocks.MockUserValidationService)

	svc := authSvc.NewAuthService(mockUserRepo, mockTokenRepo, mockPasswordSvc, mockJwtSvc, mockUserValSvc)

	mockTokenRepo.On("ValidateRefreshToken", mock.Anything, "bad").Return(0, errors.New("not found"))
	_, err := svc.Refresh(context.Background(), "bad")
	assert.ErrorIs(t, err, authSvc.ErrInvalidRefreshToken)
}

func TestAuthService_Logout(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockTokenRepo := new(mocks.MockTokenRepository)
	mockPasswordSvc := new(svcMocks.MockPasswordService)
	mockJwtSvc := new(svcMocks.MockJWTService)
	mockUserValSvc := new(svcMocks.MockUserValidationService)

	svc := authSvc.NewAuthService(mockUserRepo, mockTokenRepo, mockPasswordSvc, mockJwtSvc, mockUserValSvc)

	mockTokenRepo.On("DeleteRefreshToken", mock.Anything, "refresh").Return(nil)
	err := svc.Logout(context.Background(), "", "refresh")
	assert.NoError(t, err)

	claims := &authSvc.Claims{RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}
	mockJwtSvc.On("ValidateAccessToken", "access").Return(claims, nil)
	mockTokenRepo.On("BlacklistAccessToken", mock.Anything, "access", mock.Anything).Return(nil)
	err = svc.Logout(context.Background(), "access", "")
	assert.NoError(t, err)
}

func TestAuthService_ValidateAccessToken(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockTokenRepo := new(mocks.MockTokenRepository)
	mockPasswordSvc := new(svcMocks.MockPasswordService)
	mockJwtSvc := new(svcMocks.MockJWTService)
	mockUserValSvc := new(svcMocks.MockUserValidationService)

	svc := authSvc.NewAuthService(mockUserRepo, mockTokenRepo, mockPasswordSvc, mockJwtSvc, mockUserValSvc)

	mockTokenRepo.On("IsAccessTokenBlacklisted", mock.Anything, "blacklisted").Return(true, nil)
	_, err := svc.ValidateAccessToken(context.Background(), "blacklisted")
	assert.ErrorIs(t, err, authSvc.ErrTokenBlacklisted)

	mockTokenRepo.On("IsAccessTokenBlacklisted", mock.Anything, "valid").Return(false, nil)
	claims := &authSvc.Claims{UserID: 1}
	mockJwtSvc.On("ValidateAccessToken", "valid").Return(claims, nil)
	user := &entity.User{ID: 1}
	mockUserRepo.On("GetByID", mock.Anything, 1).Return(user, nil)
	result, err := svc.ValidateAccessToken(context.Background(), "valid")
	assert.NoError(t, err)
	assert.Equal(t, user, result)
}

func TestAuthService_ChangePassword(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockTokenRepo := new(mocks.MockTokenRepository)
	mockPasswordSvc := new(svcMocks.MockPasswordService)
	mockJwtSvc := new(svcMocks.MockJWTService)
	mockUserValSvc := new(svcMocks.MockUserValidationService)

	svc := authSvc.NewAuthService(mockUserRepo, mockTokenRepo, mockPasswordSvc, mockJwtSvc, mockUserValSvc)

	user := &entity.User{ID: 1, PasswordHash: "oldHash", IsActive: true}
	mockUserRepo.On("GetByID", mock.Anything, 1).Return(user, nil)
	mockPasswordSvc.On("Verify", "oldHash", "oldPass").Return(true)
	mockUserValSvc.On("ValidatePassword", "newPass").Return(nil)
	mockPasswordSvc.On("Hash", "newPass").Return("newHash", nil)
	mockUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool { return u.PasswordHash == "newHash" })).Return(nil)
	mockTokenRepo.On("DeleteAllUserTokens", mock.Anything, 1).Return(nil)

	err := svc.ChangePassword(context.Background(), 1, "oldPass", "newPass")
	assert.NoError(t, err)
}
