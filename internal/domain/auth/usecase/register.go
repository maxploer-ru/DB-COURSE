package usecase

import (
	"ZVideo/internal/domain/auth/entity"
	repository2 "ZVideo/internal/domain/auth/repository"
	"ZVideo/internal/domain/auth/service"
	"context"
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

type RegisterUserUseCase struct {
	userRepo    repository2.UserRepository
	roleRepo    repository2.RoleRepository
	passwordSvc service.PasswordService
	userValSvc  service.UserValidationService
	authSvc     service.AuthService
}

func NewRegisterUserUseCase(
	userRepo repository2.UserRepository,
	roleRepo repository2.RoleRepository,
	passwordSvc service.PasswordService,
	userValSvc service.UserValidationService,
	authSvc service.AuthService,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		passwordSvc: passwordSvc,
		userValSvc:  userValSvc,
		authSvc:     authSvc,
	}
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, cmd RegisterUserCommand) (*RegisterUserResult, error) {
	if err := uc.userValSvc.ValidateNewUser(ctx, cmd.Email, cmd.Username, cmd.Password); err != nil {
		return nil, err
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
		Role:                 defaultRole,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	authResult, err := uc.authSvc.Login(ctx, cmd.Email, cmd.Password)
	if err != nil {
		return nil, err
	}

	return &RegisterUserResult{
		User:         user,
		AccessToken:  authResult.AccessToken,
		RefreshToken: authResult.RefreshToken,
	}, nil
}
