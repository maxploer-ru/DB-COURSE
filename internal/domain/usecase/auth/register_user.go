package auth

import (
	"ZVideo/internal/domain/entity"
	"ZVideo/internal/domain/repository"
	"ZVideo/internal/pkg/jwt"
	"ZVideo/internal/pkg/password"
	"context"
	"errors"
	"time"
)

type RegisterUserCommand struct {
	Username string
	Email    string
	Password string
}

type RegisterUserResult struct {
	User         *entity.User
	AccessToken  string
	RefreshToken string
}

var (
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrWeakPassword          = errors.New("password is too weak")
)

type RegisterUserUseCase struct {
	userRepo    repository.UserRepository
	roleRepo    repository.RoleRepository
	passwordSvc password.Service
	jwtSvc      jwt.Service
}

func NewRegisterUserUseCase(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	passwordSvc password.Service,
	jwtSvc jwt.Service,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		passwordSvc: passwordSvc,
		jwtSvc:      jwtSvc,
	}
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, cmd RegisterUserCommand) (*RegisterUserResult, error) {
	existing, _ := uc.userRepo.GetByEmail(ctx, cmd.Email)
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	existing, _ = uc.userRepo.GetByUsername(ctx, cmd.Username)
	if existing != nil {
		return nil, ErrUsernameAlreadyExists
	}

	if !uc.passwordSvc.IsStrong(cmd.Password) {
		return nil, ErrWeakPassword
	}

	hashedPassword, err := uc.passwordSvc.Hash(cmd.Password)
	if err != nil {
		return nil, err
	}

	defaultRole, err := uc.roleRepo.GetDefaultRole(ctx)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		RoleID:               defaultRole.ID,
		Username:             cmd.Username,
		Email:                cmd.Email,
		PasswordHash:         hashedPassword,
		NotificationsEnabled: true,
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		Role:                 defaultRole,
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	accessToken, err := uc.jwtSvc.GenerateAccessToken(user.ID, user.Username, defaultRole.Name)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.jwtSvc.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &RegisterUserResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
