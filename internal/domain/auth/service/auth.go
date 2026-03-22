package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ZVideo/internal/domain/auth/entity"
	"ZVideo/internal/domain/auth/repository"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrTokenBlacklisted    = errors.New("token is blacklisted")
	ErrUserInactive        = errors.New("user is inactive")
	ErrUserNotFound        = errors.New("user not found")
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (*AuthResult, error)

	Refresh(ctx context.Context, refreshToken string) (*AuthResult, error)

	Logout(ctx context.Context, accessToken, refreshToken string) error

	ValidateAccessToken(ctx context.Context, token string) (*entity.User, error)
}

type AuthResult struct {
	User         *entity.User
	AccessToken  string
	RefreshToken string
}

type authService struct {
	userRepo    repository.UserRepository
	tokenRepo   repository.TokenRepository
	passwordSvc PasswordService
	jwtSvc      JWTService
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	passwordSvc PasswordService,
	jwtSvc JWTService,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		passwordSvc: passwordSvc,
		jwtSvc:      jwtSvc,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !s.passwordSvc.Verify(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	accessToken, err := s.jwtSvc.GenerateAccessToken(user.ID, user.Username, user.Role.Name)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.jwtSvc.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.tokenRepo.SaveRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	userID, err := s.tokenRepo.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	accessToken, err := s.jwtSvc.GenerateAccessToken(user.ID, user.Username, user.Role.Name)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := s.jwtSvc.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	_ = s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
	if err := s.tokenRepo.SaveRefreshToken(ctx, user.ID, newRefreshToken); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *authService) Logout(ctx context.Context, accessToken, refreshToken string) error {

	if refreshToken != "" {
		_ = s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
	}

	if accessToken != "" {

		claims, err := s.jwtSvc.ValidateAccessToken(accessToken)
		if err == nil {
			ttl := time.Until(claims.ExpiresAt.Time)
			if ttl > 0 {
				_ = s.tokenRepo.BlacklistAccessToken(ctx, accessToken, ttl)
			}
		}
	}

	return nil
}

func (s *authService) ValidateAccessToken(ctx context.Context, token string) (*entity.User, error) {

	blacklisted, err := s.tokenRepo.IsAccessTokenBlacklisted(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("check blacklist: %w", err)
	}
	if blacklisted {
		return nil, ErrTokenBlacklisted
	}

	claims, err := s.jwtSvc.ValidateAccessToken(token)
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}
