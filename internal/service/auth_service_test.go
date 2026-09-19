package service_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"ZVideo/internal/testing/mocks"
	"ZVideo/internal/testing/mother"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type AuthServiceTestSuite struct {
	suite.Suite
	mockUserRepo    *mocks.UserRepository
	mockRoleRepo    *mocks.RoleRepository
	mockRefreshRepo *mocks.RefreshSessionRepository
	mockPwdSvc      *mocks.PasswordService
	mockJwtSvc      *mocks.JWTService
	mockUserValSvc  *mocks.UserValidatorService
	service         service.AuthService
	userMother      mother.UserMother
	roleMother      mother.RoleMother
}

func (s *AuthServiceTestSuite) SetupTest() {
	s.mockUserRepo = mocks.NewUserRepository(s.T())
	s.mockRoleRepo = mocks.NewRoleRepository(s.T())
	s.mockRefreshRepo = mocks.NewRefreshSessionRepository(s.T())
	s.mockPwdSvc = mocks.NewPasswordService(s.T())
	s.mockJwtSvc = mocks.NewJWTService(s.T())
	s.mockUserValSvc = mocks.NewUserValidatorService(s.T())

	s.service = service.NewAuthService(
		s.mockUserRepo,
		s.mockRoleRepo,
		s.mockRefreshRepo,
		s.mockPwdSvc,
		s.mockJwtSvc,
		s.mockUserValSvc,
	)
	s.userMother = mother.UserMother{}
	s.roleMother = mother.RoleMother{}
}

func (s *AuthServiceTestSuite) TestLogin_Negative_UserNotFound_EquivalencePartitioning() {
	ctx := context.Background()
	email := "test@example.com"
	password := "password123"

	s.mockUserRepo.On("GetByEmail", ctx, email).Return(nil, nil)

	res, err := s.service.Login(ctx, email, password)

	s.ErrorIs(err, domain.ErrInvalidUserCredentials)
	s.Nil(res)
}

func (s *AuthServiceTestSuite) TestLogin_Positive_Success_Combinatorial() {
	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	user := s.userMother.ValidActiveUser()
	user.Role = s.roleMother.DefaultUserRole()

	tokenData := &domain.AccessTokenData{
		UserID:   user.ID,
		UserName: user.Username,
		Role:     user.Role.Name,
	}
	refreshData := &domain.RefreshTokenData{
		UserID:    user.ID,
		TokenID:   "token_id_123",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	s.mockUserRepo.On("GetByEmail", ctx, email).Return(user, nil)
	s.mockPwdSvc.On("ComparePassword", ctx, password, user.PasswordHash).Return(nil)
	s.mockJwtSvc.On("GenerateAccessToken", ctx, tokenData).Return("access_token", nil)
	s.mockJwtSvc.On("GenerateRefreshToken", ctx, user.ID).Return("refresh_token", refreshData, nil)
	s.mockRefreshRepo.On("Save", ctx, refreshData.TokenID, user.ID, refreshData.ExpiresAt).Return(nil)

	res, err := s.service.Login(ctx, email, password)

	s.NoError(err)
	s.NotNil(res)
	s.Equal("access_token", res.AccessToken)
	s.Equal("refresh_token", res.RefreshToken)
	s.mockRefreshRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestRefresh_Negative_InvalidToken_EquivalencePartitioning() {
	ctx := context.Background()
	refreshToken := "invalid_token"

	s.mockJwtSvc.On("ValidateRefreshToken", ctx, refreshToken).Return(nil, domain.ErrInvalidRefreshToken)

	res, err := s.service.Refresh(ctx, refreshToken)

	s.ErrorIs(err, domain.ErrInvalidRefreshToken)
	s.Nil(res)
}

func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
