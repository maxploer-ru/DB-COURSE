package usecase

import (
	"context"
	"errors"
	"fmt"

	"ZVideo/internal/domain/auth/service"
)

type LogoutCommand struct {
	AccessToken  string
	RefreshToken string
	UserID       int
}

var (
	ErrInvalidToken = errors.New("invalid token")
)

type LogoutUseCase struct {
	authSvc service.AuthService
}

func NewLogoutUseCase(authSvc service.AuthService) *LogoutUseCase {
	return &LogoutUseCase{
		authSvc: authSvc,
	}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, cmd LogoutCommand) error {
	if cmd.AccessToken == "" && cmd.RefreshToken == "" {
		return fmt.Errorf("no tokens provided")
	}

	if err := uc.authSvc.Logout(ctx, cmd.AccessToken, cmd.RefreshToken); err != nil {

		if errors.Is(err, service.ErrInvalidRefreshToken) {
			return ErrInvalidToken
		}
		return fmt.Errorf("logout failed: %w", err)
	}

	return nil
}
