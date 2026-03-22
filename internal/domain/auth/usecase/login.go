package usecase

import (
	"context"
	"errors"
	"fmt"

	"ZVideo/internal/domain/auth/entity"
	"ZVideo/internal/domain/auth/service"
)

type LoginUserCommand struct {
	Email    string
	Password string
}

type LoginUserResult struct {
	User         *entity.User
	AccessToken  string
	RefreshToken string
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountDeactivated = errors.New("account is deactivated")
)

type LoginUserUseCase struct {
	authSvc service.AuthService
}

func NewLoginUserUseCase(authSvc service.AuthService) *LoginUserUseCase {
	return &LoginUserUseCase{
		authSvc: authSvc,
	}
}

func (uc *LoginUserUseCase) Execute(ctx context.Context, cmd LoginUserCommand) (*LoginUserResult, error) {
	if cmd.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if cmd.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	result, err := uc.authSvc.Login(ctx, cmd.Email, cmd.Password)
	if err != nil {

		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			return nil, ErrInvalidCredentials
		case errors.Is(err, service.ErrUserInactive):
			return nil, ErrAccountDeactivated
		default:
			return nil, fmt.Errorf("login failed: %w", err)
		}
	}

	return &LoginUserResult{
		User:         result.User,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}
