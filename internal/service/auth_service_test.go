package service_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"ZVideo/internal/testing/mocks"
	"ZVideo/internal/testing/mother"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
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

func (s *AuthServiceTestSuite) TestRegister_Positive_StateTransition() {
	ctx := context.Background()
	user := s.userMother.ValidActiveUser()
	role := s.roleMother.DefaultUserRole()

	s.mockUserValSvc.On("ValidateNewUser", ctx, "test@example.com", "test_user", "password123").Return(nil)
	s.mockUserRepo.On("ExistsByUsername", ctx, "test_user").Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", ctx, "test@example.com").Return(false, nil)
	s.mockRoleRepo.On("GetDefaultRole", ctx).Return(role, nil)
	s.mockPwdSvc.On("HashPassword", ctx, "password123").Return("hashed", nil)
	s.mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Run(func(args mock.Arguments) {
		args.Get(1).(*domain.User).ID = user.ID
	}).Return(nil)

	created, err := s.service.Register(ctx, "test_user", "test@example.com", "password123")

	s.NoError(err)
	s.NotNil(created)
	s.Equal(user.ID, created.ID)
	s.Empty(created.PasswordHash)
}

func (s *AuthServiceTestSuite) TestRegister_Negative_InvalidUser_EquivalencePartitioning() {
	ctx := context.Background()
	validationErr := domain.ErrInvalidUserCredentials

	s.mockUserValSvc.On("ValidateNewUser", ctx, "bad@example.com", "bad", "short").Return(validationErr)

	created, err := s.service.Register(ctx, "bad", "bad@example.com", "short")

	s.ErrorIs(err, validationErr)
	s.Nil(created)
}

func (s *AuthServiceTestSuite) TestRefresh_Negative_InvalidToken_EquivalencePartitioning() {
	ctx := context.Background()
	refreshToken := "invalid_token"

	s.mockJwtSvc.On("ValidateRefreshToken", ctx, refreshToken).Return(nil, domain.ErrInvalidRefreshToken)

	res, err := s.service.Refresh(ctx, refreshToken)

	s.ErrorIs(err, domain.ErrInvalidRefreshToken)
	s.Nil(res)
}

func (s *AuthServiceTestSuite) TestRefresh_Positive_StateTransition() {
	ctx := context.Background()
	user := s.userMother.ValidActiveUser()
	user.Role = s.roleMother.DefaultUserRole()
	refreshData := &domain.RefreshTokenData{UserID: user.ID, TokenID: "old", ExpiresAt: time.Now().Add(time.Hour)}
	newRefreshData := &domain.RefreshTokenData{UserID: user.ID, TokenID: "new", ExpiresAt: time.Now().Add(2 * time.Hour)}
	accessData := &domain.AccessTokenData{UserID: user.ID, UserName: user.Username, Role: user.Role.Name}

	s.mockJwtSvc.On("ValidateRefreshToken", ctx, "refresh").Return(refreshData, nil)
	s.mockRefreshRepo.On("GetUserID", ctx, refreshData.TokenID).Return(user.ID, true, nil)
	s.mockUserRepo.On("GetByID", ctx, user.ID).Return(user, nil)
	s.mockJwtSvc.On("GenerateAccessToken", ctx, accessData).Return("access", nil)
	s.mockJwtSvc.On("GenerateRefreshToken", ctx, user.ID).Return("new-refresh", newRefreshData, nil)
	s.mockRefreshRepo.On("Rotate", ctx, refreshData.TokenID, newRefreshData.TokenID, user.ID, newRefreshData.ExpiresAt).Return(true, nil)

	result, err := s.service.Refresh(ctx, "refresh")

	s.NoError(err)
	s.Equal("access", result.AccessToken)
	s.Equal("new-refresh", result.RefreshToken)
}

func (s *AuthServiceTestSuite) TestLogout_Positive_StateTransition() {
	ctx := context.Background()
	accessData := &domain.AccessTokenData{UserID: 1, UserName: "test_user", Role: domain.RoleUser}
	refreshData := &domain.RefreshTokenData{UserID: 1, TokenID: "refresh-id", ExpiresAt: time.Now().Add(time.Hour)}

	s.mockJwtSvc.On("ValidateAccessToken", ctx, "access").Return(accessData, nil)
	s.mockJwtSvc.On("ValidateRefreshToken", ctx, "refresh").Return(refreshData, nil)
	s.mockRefreshRepo.On("Delete", ctx, refreshData.TokenID).Return(nil)

	err := s.service.Logout(ctx, "access", "refresh")

	s.NoError(err)
}

func (s *AuthServiceTestSuite) TestLogout_Negative_MismatchedUsers_Combinatorial() {
	ctx := context.Background()
	accessData := &domain.AccessTokenData{UserID: 1}
	refreshData := &domain.RefreshTokenData{UserID: 2, TokenID: "refresh-id"}

	s.mockJwtSvc.On("ValidateAccessToken", ctx, "access").Return(accessData, nil)
	s.mockJwtSvc.On("ValidateRefreshToken", ctx, "refresh").Return(refreshData, nil)

	err := s.service.Logout(ctx, "access", "refresh")

	s.ErrorIs(err, domain.ErrForbidden)
}

func (s *AuthServiceTestSuite) TestValidateAccessToken_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	user := s.userMother.ValidActiveUser()
	user.Role = s.roleMother.DefaultUserRole()
	tokenData := &domain.AccessTokenData{UserID: user.ID, UserName: "stale", Role: "stale"}

	s.mockJwtSvc.On("ValidateAccessToken", ctx, "access").Return(tokenData, nil)
	s.mockUserRepo.On("GetByID", ctx, user.ID).Return(user, nil)

	validated, err := s.service.ValidateAccessToken(ctx, "access")

	s.NoError(err)
	s.Equal(user.Username, validated.UserName)
	s.Equal(user.Role.Name, validated.Role)
}

func (s *AuthServiceTestSuite) TestValidateAccessToken_Negative_BannedUser_EquivalencePartitioning() {
	ctx := context.Background()
	user := s.userMother.BannedUser()
	tokenData := &domain.AccessTokenData{UserID: user.ID}

	s.mockJwtSvc.On("ValidateAccessToken", ctx, "access").Return(tokenData, nil)
	s.mockUserRepo.On("GetByID", ctx, user.ID).Return(user, nil)

	validated, err := s.service.ValidateAccessToken(ctx, "access")

	s.ErrorIs(err, domain.ErrUserIsBanned)
	s.Nil(validated)
}

func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
