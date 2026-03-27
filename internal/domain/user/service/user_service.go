package service

import (
	"ZVideo/internal/domain/user/repository"
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
)

type Profile struct {
	ID                   int    `json:"id"`
	Username             string `json:"username"`
	Email                string `json:"email"`
	NotificationsEnabled bool   `json:"notifications_enabled"`
	Role                 string `json:"role"`
	IsActive             bool   `json:"is_active"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}

type UpdateProfileCommand struct {
	Username             *string
	Email                *string
	NotificationsEnabled *bool
}

type AuthService interface {
	ChangePassword(ctx context.Context, userID int, oldPassword, newPassword string) error
	RevokeAllUserTokens(ctx context.Context, userID int) error
}

type UserService interface {
	GetProfile(ctx context.Context, userID int) (*Profile, error)
	UpdateProfile(ctx context.Context, userID int, cmd UpdateProfileCommand) (*Profile, error)
}

type userService struct {
	userRepo repository.UserRepository
	authSvc  AuthService
}

func NewUserService(userRepo repository.UserRepository, authSvc AuthService) UserService {
	return &userService{
		userRepo: userRepo,
		authSvc:  authSvc,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID int) (*Profile, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	role := ""
	if user.Role != nil {
		role = user.Role.Name
	}

	return &Profile{
		ID:                   user.ID,
		Username:             user.Username,
		Email:                user.Email,
		NotificationsEnabled: user.NotificationsEnabled,
		Role:                 role,
		IsActive:             user.IsActive,
		CreatedAt:            user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID int, cmd UpdateProfileCommand) (*Profile, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if cmd.Username != nil && *cmd.Username != user.Username {
		exists, err := s.userRepo.ExistsByUsername(ctx, *cmd.Username)
		if err != nil {
			return nil, fmt.Errorf("check username: %w", err)
		}
		if exists {
			return nil, ErrUsernameAlreadyExists
		}
		user.Username = *cmd.Username
	}

	if cmd.Email != nil && *cmd.Email != user.Email {
		exists, err := s.userRepo.ExistsByEmail(ctx, *cmd.Email)
		if err != nil {
			return nil, fmt.Errorf("check email: %w", err)
		}
		if exists {
			return nil, ErrEmailAlreadyExists
		}
		user.Email = *cmd.Email
	}

	if cmd.NotificationsEnabled != nil {
		user.NotificationsEnabled = *cmd.NotificationsEnabled
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return s.GetProfile(ctx, userID)
}
