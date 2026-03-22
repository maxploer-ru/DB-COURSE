package usecase

import (
	"context"
	"errors"
	"fmt"

	"ZVideo/internal/domain/auth/service"
)

type RefreshTokenCommand struct {
	RefreshToken string
}

type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
}

var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type RefreshTokenUseCase struct {
	authSvc service.AuthService
}

func NewRefreshTokenUseCase(authSvc service.AuthService) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		authSvc: authSvc,
	}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, cmd RefreshTokenCommand) (*RefreshTokenResult, error) {
	if cmd.RefreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	result, err := uc.authSvc.Refresh(ctx, cmd.RefreshToken)
	if err != nil {

		if errors.Is(err, service.ErrInvalidRefreshToken) {
			return nil, ErrInvalidRefreshToken
		}
		if errors.Is(err, service.ErrUserNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("refresh failed: %w", err)
	}

	return &RefreshTokenResult{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}
