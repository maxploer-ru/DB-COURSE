package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type AuthService interface {
	Register(ctx context.Context, username, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (*AuthResult, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
	ValidateAccessToken(ctx context.Context, token string) (*domain.AccessTokenData, error)
}

type UserValidatorService interface {
	ValidateNewUser(ctx context.Context, email, nickname, password string) error
}

type PasswordService interface {
	HashPassword(ctx context.Context, password string) (string, error)
	ComparePassword(ctx context.Context, password, hash string) error
}

type JWTService interface {
	GenerateAccessToken(ctx context.Context, data *domain.AccessTokenData) (string, error)
	GenerateRefreshToken(ctx context.Context, userID int) (string, *domain.RefreshTokenData, error)
	ValidateAccessToken(ctx context.Context, accessToken string) (*domain.AccessTokenData, error)
	ValidateRefreshToken(ctx context.Context, refreshToken string) (*domain.RefreshTokenData, error)
}

type AuthResult struct {
	User             *domain.User
	AccessToken      string
	RefreshToken     string
	RefreshExpiresAt time.Time
}

type authService struct {
	userRepo    repository.UserRepository
	roleRepo    repository.RoleRepository
	refreshRepo repository.RefreshSessionRepository
	pwdSvc      PasswordService
	userValSvc  UserValidatorService
	jwtSvc      JWTService
}

func NewAuthService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	refreshRepo repository.RefreshSessionRepository,
	pwdSvc PasswordService,
	jwtSvc JWTService,
	userValSvc UserValidatorService,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		refreshRepo: refreshRepo,
		pwdSvc:      pwdSvc,
		jwtSvc:      jwtSvc,
		userValSvc:  userValSvc,
	}
}

func (s *authService) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	username = strings.TrimSpace(username)
	email = strings.ToLower(strings.TrimSpace(email))
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "AuthService"),
		slog.String("operation", "Register"),
		slog.String("username", username),
		slog.String("email", email),
	)

	logger.DebugContext(ctx, "Validating new user")
	if err := s.userValSvc.ValidateNewUser(ctx, email, username, password); err != nil {
		logger.WarnContext(ctx, "User validation failed", slog.String("error", err.Error()))
		return nil, err
	}

	logger.DebugContext(ctx, "Checking username existence")
	exists, err := s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check username existence", slog.String("error", err.Error()))
		return nil, err
	}
	if exists {
		logger.WarnContext(ctx, "Username already taken")
		return nil, domain.ErrUserNameAlreadyExists
	}

	logger.DebugContext(ctx, "Checking email existence")
	exists, err = s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check email existence", slog.String("error", err.Error()))
		return nil, err
	}
	if exists {
		logger.WarnContext(ctx, "Email already registered")
		return nil, domain.ErrUserEmailAlreadyExists
	}

	logger.DebugContext(ctx, "Getting default role")
	defaultRole, err := s.roleRepo.GetDefaultRole(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get default role", slog.String("error", err.Error()))
		return nil, err
	}

	logger.DebugContext(ctx, "Hashing password")
	passwordHash, err := s.pwdSvc.HashPassword(ctx, password)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to hash password", slog.String("error", err.Error()))
		return nil, fmt.Errorf("hash password failed: %w", err)
	}

	user := &domain.User{
		Username:             username,
		Email:                email,
		PasswordHash:         passwordHash,
		IsActive:             true,
		NotificationsEnabled: true,
		Role:                 defaultRole,
	}

	logger.DebugContext(ctx, "Creating user in repository")
	if err = s.userRepo.Create(ctx, user); err != nil {
		logger.ErrorContext(ctx, "Failed to create user", slog.String("error", err.Error()))
		return nil, fmt.Errorf("create user failed: %w", err)
	}

	logger.InfoContext(ctx, "User registered successfully", slog.Int("user_id", user.ID))
	createdUser := *user
	createdUser.PasswordHash = ""
	return &createdUser, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "AuthService"),
		slog.String("operation", "Login"),
		slog.String("email", email),
	)

	logger.DebugContext(ctx, "Fetching user by email")
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get user by email", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get user by email failed: %w", err)
	}
	if user == nil {
		logger.WarnContext(ctx, "User not found")
		return nil, domain.ErrInvalidUserCredentials
	}
	if user.Role == nil {
		logger.ErrorContext(ctx, "User has no role assigned")
		return nil, domain.ErrInternalServer
	}
	logger = logger.With(slog.Int("user_id", user.ID))

	logger.DebugContext(ctx, "Comparing password")
	if err = s.pwdSvc.ComparePassword(ctx, password, user.PasswordHash); err != nil {
		logger.WarnContext(ctx, "Invalid password")
		return nil, domain.ErrInvalidUserCredentials
	}

	if !user.IsActive {
		logger.WarnContext(ctx, "User is banned")
		return nil, domain.ErrUserIsBanned
	}

	tokenData := domain.AccessTokenData{
		UserID:   user.ID,
		UserName: user.Username,
		Role:     user.Role.Name,
	}

	logger.DebugContext(ctx, "Generating access token")
	accessToken, err := s.jwtSvc.GenerateAccessToken(ctx, &tokenData)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to generate access token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("generate access token failed: %w", err)
	}

	logger.DebugContext(ctx, "Generating refresh token")
	refreshToken, refreshData, err := s.jwtSvc.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to generate refresh token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("generate refresh token failed: %w", err)
	}
	if err := s.refreshRepo.Save(ctx, refreshData.TokenID, user.ID, refreshData.ExpiresAt); err != nil {
		logger.ErrorContext(ctx, "Failed to persist refresh session", slog.String("error", err.Error()))
		return nil, fmt.Errorf("persist refresh session failed: %w", err)
	}

	logger.InfoContext(ctx, "User logged in successfully")
	publicUser := *user
	publicUser.PasswordHash = ""
	return &AuthResult{
		User:             &publicUser,
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshData.ExpiresAt,
	}, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "AuthService"),
		slog.String("operation", "Refresh"),
	)

	logger.DebugContext(ctx, "Validating refresh token")
	refreshData, err := s.jwtSvc.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		logger.WarnContext(ctx, "Invalid refresh token", slog.String("error", err.Error()))
		return nil, domain.ErrInvalidRefreshToken
	}
	logger = logger.With(slog.Int("user_id", refreshData.UserID))

	storedUserID, found, err := s.refreshRepo.GetUserID(ctx, refreshData.TokenID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to read refresh session", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get refresh session failed: %w", err)
	}
	if !found || storedUserID != refreshData.UserID {
		logger.WarnContext(ctx, "Refresh session not found or mismatched")
		return nil, domain.ErrInvalidRefreshToken
	}

	logger.DebugContext(ctx, "Fetching user by ID")
	user, err := s.userRepo.GetByID(ctx, refreshData.UserID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get user", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get user failed: %w", err)
	}
	if user == nil {
		logger.WarnContext(ctx, "User not found")
		return nil, domain.ErrUserNotFound
	}
	if !user.IsActive {
		logger.WarnContext(ctx, "User is banned")
		return nil, domain.ErrUserIsBanned
	}
	if user.Role == nil {
		logger.ErrorContext(ctx, "User has no role assigned")
		return nil, domain.ErrInternalServer
	}

	tokenData := domain.AccessTokenData{
		UserID:   user.ID,
		UserName: user.Username,
		Role:     user.Role.Name,
	}

	logger.DebugContext(ctx, "Generating new access token")
	accessToken, err := s.jwtSvc.GenerateAccessToken(ctx, &tokenData)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to generate access token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("generate access token failed: %w", err)
	}

	logger.DebugContext(ctx, "Generating new refresh token")
	newRefreshToken, newRefreshData, err := s.jwtSvc.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to generate refresh token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("generate refresh token failed: %w", err)
	}
	rotated, err := s.refreshRepo.Rotate(ctx, refreshData.TokenID, newRefreshData.TokenID, user.ID, newRefreshData.ExpiresAt)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to rotate refresh session", slog.String("error", err.Error()))
		return nil, fmt.Errorf("rotate refresh session failed: %w", err)
	}
	if !rotated {
		logger.WarnContext(ctx, "Refresh session rotation failed because old session is missing")
		return nil, domain.ErrInvalidRefreshToken
	}

	logger.InfoContext(ctx, "Tokens refreshed successfully")
	publicUser := *user
	publicUser.PasswordHash = ""
	return &AuthResult{
		User:             &publicUser,
		AccessToken:      accessToken,
		RefreshToken:     newRefreshToken,
		RefreshExpiresAt: newRefreshData.ExpiresAt,
	}, nil
}

func (s *authService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "AuthService"),
		slog.String("operation", "Logout"),
	)
	if accessToken == "" {
		logger.WarnContext(ctx, "Logout rejected because access token is missing")
		return domain.ErrInvalidAccessToken
	}
	accessData, err := s.jwtSvc.ValidateAccessToken(ctx, accessToken)
	if err != nil {
		logger.WarnContext(ctx, "Logout rejected because access token is invalid", slog.Any("error", err))
		return domain.ErrInvalidAccessToken
	}
	if refreshToken == "" {
		logger.WarnContext(ctx, "Logout rejected because refresh token is missing", slog.Int("user_id", accessData.UserID))
		return domain.ErrInvalidRefreshToken
	}
	refreshData, err := s.jwtSvc.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		logger.WarnContext(ctx, "Logout rejected because refresh token is invalid", slog.Int("user_id", accessData.UserID), slog.Any("error", err))
		return domain.ErrInvalidRefreshToken
	}
	if refreshData.UserID != accessData.UserID {
		logger.WarnContext(ctx, "Logout rejected because token users do not match", slog.Int("access_user_id", accessData.UserID), slog.Int("refresh_user_id", refreshData.UserID))
		return domain.ErrForbidden
	}
	logger = logger.With(slog.Int("user_id", accessData.UserID))
	if err := s.refreshRepo.Delete(ctx, refreshData.TokenID); err != nil {
		logger.ErrorContext(ctx, "Failed to delete refresh session", slog.String("error", err.Error()))
		return fmt.Errorf("delete refresh session failed: %w", err)
	}
	logger.InfoContext(ctx, "Refresh session revoked successfully")
	return nil
}

func (s *authService) ValidateAccessToken(ctx context.Context, token string) (*domain.AccessTokenData, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "AuthService"),
		slog.String("operation", "ValidateAccessToken"),
	)

	tokenData, err := s.jwtSvc.ValidateAccessToken(ctx, token)
	if err != nil {
		logger.WarnContext(ctx, "Invalid access token", slog.String("error", err.Error()))
		return nil, domain.ErrInvalidAccessToken
	}

	user, err := s.userRepo.GetByID(ctx, tokenData.UserID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to load authenticated user", slog.Int("user_id", tokenData.UserID), slog.Any("error", err))
		return nil, fmt.Errorf("load authenticated user: %w", err)
	}
	if user == nil {
		logger.WarnContext(ctx, "Authenticated user not found", slog.Int("user_id", tokenData.UserID))
		return nil, domain.ErrInvalidAccessToken
	}
	if !user.IsActive {
		logger.WarnContext(ctx, "Authenticated user is banned", slog.Int("user_id", user.ID))
		return nil, domain.ErrUserIsBanned
	}
	if user.Role == nil {
		logger.ErrorContext(ctx, "Authenticated user has no role", slog.Int("user_id", user.ID))
		return nil, domain.ErrInternalServer
	}
	// Mutable authorization state is authoritative in the database.
	tokenData.Role = user.Role.Name
	tokenData.UserName = user.Username
	return tokenData, nil
}
