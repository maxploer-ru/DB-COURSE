package service_test

import (
	"ZVideo/internal/domain"
	service "ZVideo/internal/service"
	"ZVideo/mocks"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	roleRepo := mocks.NewRoleRepository(t)
	refreshRepo := mocks.NewRefreshSessionRepository(t)
	pwdSvc := mocks.NewPasswordService(t)
	jwtSvc := mocks.NewJWTService(t)
	userValSvc := mocks.NewUserValidatorService(t)

	role := &domain.Role{ID: 1, Name: "user"}
	userValSvc.On("ValidateNewUser", ctx, "u@example.com", "nick", "pass").Return(nil)
	userRepo.On("ExistsByUsername", ctx, "nick").Return(false, nil)
	userRepo.On("ExistsByEmail", ctx, "u@example.com").Return(false, nil)
	roleRepo.On("GetDefaultRole", ctx).Return(role, nil)
	pwdSvc.On("HashPassword", ctx, "pass").Return("hash", nil)
	userRepo.On("Create", ctx, mock.MatchedBy(func(u *domain.User) bool {
		return u.Username == "nick" && u.Email == "u@example.com" && u.PasswordHash == "hash" && u.Role == role
	})).Return(nil)

	svc := service.NewAuthService(userRepo, roleRepo, refreshRepo, pwdSvc, jwtSvc, userValSvc)
	err := svc.Register(ctx, "nick", "u@example.com", "pass")
	require.NoError(t, err)
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	roleRepo := mocks.NewRoleRepository(t)
	refreshRepo := mocks.NewRefreshSessionRepository(t)
	pwdSvc := mocks.NewPasswordService(t)
	jwtSvc := mocks.NewJWTService(t)
	userValSvc := mocks.NewUserValidatorService(t)

	role := &domain.Role{ID: 1, Name: "user"}
	user := &domain.User{ID: 7, Username: "nick", Email: "u@example.com", PasswordHash: "hash", IsActive: true, Role: role}
	refreshData := &domain.RefreshTokenData{TokenID: "token-1", UserID: 7, ExpiresAt: time.Now().Add(time.Hour)}

	userRepo.On("GetByEmail", ctx, "u@example.com").Return(user, nil)
	pwdSvc.On("ComparePassword", ctx, "pass", "hash").Return(nil)
	jwtSvc.On("GenerateAccessToken", ctx, mock.Anything).Return("access", nil)
	jwtSvc.On("GenerateRefreshToken", ctx, 7).Return("refresh", refreshData, nil)
	refreshRepo.On("Save", ctx, refreshData.TokenID, 7, refreshData.ExpiresAt).Return(nil)

	svc := service.NewAuthService(userRepo, roleRepo, refreshRepo, pwdSvc, jwtSvc, userValSvc)
	result, err := svc.Login(ctx, "u@example.com", "pass")
	require.NoError(t, err)
	require.Equal(t, "access", result.AccessToken)
	require.Equal(t, "refresh", result.RefreshToken)
}

func TestAuthService_Refresh(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	roleRepo := mocks.NewRoleRepository(t)
	refreshRepo := mocks.NewRefreshSessionRepository(t)
	pwdSvc := mocks.NewPasswordService(t)
	jwtSvc := mocks.NewJWTService(t)
	userValSvc := mocks.NewUserValidatorService(t)

	refreshData := &domain.RefreshTokenData{TokenID: "old", UserID: 3, ExpiresAt: time.Now().Add(time.Hour)}
	newRefreshData := &domain.RefreshTokenData{TokenID: "new", UserID: 3, ExpiresAt: time.Now().Add(2 * time.Hour)}
	user := &domain.User{ID: 3, Username: "nick", IsActive: true, Role: &domain.Role{Name: "user"}}

	jwtSvc.On("ValidateRefreshToken", ctx, "refresh").Return(refreshData, nil)
	refreshRepo.On("GetUserID", ctx, "old").Return(3, true, nil)
	userRepo.On("GetByID", ctx, 3).Return(user, nil)
	jwtSvc.On("GenerateAccessToken", ctx, mock.Anything).Return("access", nil)
	jwtSvc.On("GenerateRefreshToken", ctx, 3).Return("new-refresh", newRefreshData, nil)
	refreshRepo.On("Rotate", ctx, "old", "new", 3, newRefreshData.ExpiresAt).Return(true, nil)

	svc := service.NewAuthService(userRepo, roleRepo, refreshRepo, pwdSvc, jwtSvc, userValSvc)
	result, err := svc.Refresh(ctx, "refresh")
	require.NoError(t, err)
	require.Equal(t, "access", result.AccessToken)
	require.Equal(t, "new-refresh", result.RefreshToken)
}

func TestAuthService_Logout(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	roleRepo := mocks.NewRoleRepository(t)
	refreshRepo := mocks.NewRefreshSessionRepository(t)
	pwdSvc := mocks.NewPasswordService(t)
	jwtSvc := mocks.NewJWTService(t)
	userValSvc := mocks.NewUserValidatorService(t)

	refreshData := &domain.RefreshTokenData{TokenID: "token-1", UserID: 5, ExpiresAt: time.Now().Add(time.Hour)}
	jwtSvc.On("ValidateRefreshToken", ctx, "refresh").Return(refreshData, nil)
	refreshRepo.On("Delete", ctx, "token-1").Return(nil)

	svc := service.NewAuthService(userRepo, roleRepo, refreshRepo, pwdSvc, jwtSvc, userValSvc)
	err := svc.Logout(ctx, "access", "refresh")
	require.NoError(t, err)
}

func TestAuthService_ValidateAccessToken(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	roleRepo := mocks.NewRoleRepository(t)
	refreshRepo := mocks.NewRefreshSessionRepository(t)
	pwdSvc := mocks.NewPasswordService(t)
	jwtSvc := mocks.NewJWTService(t)
	userValSvc := mocks.NewUserValidatorService(t)

	jwtSvc.On("ValidateAccessToken", ctx, "token").Return(&domain.AccessTokenData{UserID: 9}, nil)
	userRepo.On("GetByID", ctx, 9).Return(&domain.User{ID: 9, Username: "u", IsActive: true, Role: &domain.Role{Name: "user"}}, nil)

	svc := service.NewAuthService(userRepo, roleRepo, refreshRepo, pwdSvc, jwtSvc, userValSvc)
	data, err := svc.ValidateAccessToken(ctx, "token")
	require.NoError(t, err)
	require.Equal(t, 9, data.UserID)
}

func TestAuthService_GetMe(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	roleRepo := mocks.NewRoleRepository(t)
	refreshRepo := mocks.NewRefreshSessionRepository(t)
	pwdSvc := mocks.NewPasswordService(t)
	jwtSvc := mocks.NewJWTService(t)
	userValSvc := mocks.NewUserValidatorService(t)

	userRepo.On("GetByID", ctx, 4).Return(&domain.User{ID: 4, Username: "me", IsActive: true, Role: &domain.Role{Name: "user"}}, nil)

	svc := service.NewAuthService(userRepo, roleRepo, refreshRepo, pwdSvc, jwtSvc, userValSvc)
	user, err := svc.GetMe(ctx, 4)
	require.NoError(t, err)
	require.Equal(t, 4, user.ID)
}

func TestAuthService_SetNotificationsEnabled(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	roleRepo := mocks.NewRoleRepository(t)
	refreshRepo := mocks.NewRefreshSessionRepository(t)
	pwdSvc := mocks.NewPasswordService(t)
	jwtSvc := mocks.NewJWTService(t)
	userValSvc := mocks.NewUserValidatorService(t)

	userRepo.On("SetNotificationsEnabled", ctx, 6, true).Return(nil)
	userRepo.On("GetByID", ctx, 6).Return(&domain.User{ID: 6, Username: "u", IsActive: true, Role: &domain.Role{Name: "user"}}, nil)

	svc := service.NewAuthService(userRepo, roleRepo, refreshRepo, pwdSvc, jwtSvc, userValSvc)
	user, err := svc.SetNotificationsEnabled(ctx, 6, true)
	require.NoError(t, err)
	require.Equal(t, 6, user.ID)
}
