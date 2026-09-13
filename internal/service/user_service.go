package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"fmt"
	"log/slog"
)

type UserService interface {
	GetProfile(ctx context.Context, userID int) (*domain.User, error)
	SetNotificationsSettings(ctx context.Context, userID int, enabled bool) error
	DeleteAccount(ctx context.Context, userID int) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID int) (*domain.User, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "UserService"),
		slog.String("operation", "GetProfile"),
		slog.Int("user_id", userID),
	)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get user", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get user failed: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}
	if !user.IsActive {
		return nil, domain.ErrUserIsBanned
	}

	return user, nil
}

func (s *userService) SetNotificationsSettings(ctx context.Context, userID int, enabled bool) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "UserService"),
		slog.String("operation", "SetNotificationsSettings"),
		slog.Int("user_id", userID),
		slog.Bool("enabled", enabled),
	)

	if err := s.userRepo.SetNotificationsEnabled(ctx, userID, enabled); err != nil {
		logger.ErrorContext(ctx, "Failed to update notifications", slog.String("error", err.Error()))
		return fmt.Errorf("update notifications failed: %w", err)
	}
	return nil
}

func (s *userService) DeleteAccount(ctx context.Context, userID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "UserService"),
		slog.String("operation", "DeleteAccount"),
		slog.Int("user_id", userID),
	)

	if err := s.userRepo.Delete(ctx, userID); err != nil {
		logger.ErrorContext(ctx, "Failed to delete user from repository", slog.String("error", err.Error()))
		return fmt.Errorf("delete user failed: %w", err)
	}

	logger.InfoContext(ctx, "Account deleted successfully")
	return nil
}
