package service

import (
	"ZVideo/internal/domain/auth/repository"
	"ZVideo/internal/infrastructure/config"
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrWeakPassword          = errors.New("password is too weak")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrForbiddenUsername     = errors.New("username is forbidden")
)

type UserValidationService interface {
	ValidateNewUser(ctx context.Context, email, username, password string) error
}

type userValidationService struct {
	userRepo repository.UserRepository
	cfg      config.PasswordConfig
}

func NewUserValidationService(
	userRepo repository.UserRepository,
	cfg config.PasswordConfig,
) UserValidationService {
	return &userValidationService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *userValidationService) ValidateNewUser(
	ctx context.Context,
	email, username, password string,
) error {

	if err := s.validatePassword(password); err != nil {
		return err
	}

	if err := s.validateEmail(email); err != nil {
		return err
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("check email: %w", err)
	}
	if exists {
		return ErrEmailAlreadyExists
	}

	exists, err = s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("check username: %w", err)
	}
	if exists {
		return ErrUsernameAlreadyExists
	}

	if s.isForbiddenUsername(username) {
		return ErrForbiddenUsername
	}

	return nil
}

func (s *userValidationService) validatePassword(password string) error {

	if len(password) < s.cfg.MinLength {
		return ErrWeakPassword
	}

	return nil
}

func (s *userValidationService) validateEmail(email string) error {
	if !strings.Contains(email, "@") {
		return ErrInvalidEmail
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ErrInvalidEmail
	}

	if parts[0] == "" || parts[1] == "" {
		return ErrInvalidEmail
	}

	return nil
}

func (s *userValidationService) isForbiddenUsername(username string) bool {
	forbidden := []string{"admin", "root", "system", "moderator", "support"}

	for _, f := range forbidden {
		if strings.EqualFold(username, f) {
			return true
		}
	}
	return false
}
