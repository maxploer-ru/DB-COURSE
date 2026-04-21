package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"log/slog"
)

type AdminService interface {
	BanUser(ctx context.Context, userID int) error
	UnbanUser(ctx context.Context, userID int) error
}

type adminService struct {
	userRepo repository.UserRepository
}

func NewAdminService(
	userRepo repository.UserRepository,
) AdminService {
	return &adminService{
		userRepo: userRepo,
	}
}

func (s *adminService) BanUser(ctx context.Context, userID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "AuthService"),
		slog.String("operation", "BanUser"),
		slog.Int("target_user_id", userID),
	)

	logger.DebugContext(ctx, "Banning user")
	if err := s.userRepo.Ban(ctx, userID); err != nil {
		logger.ErrorContext(ctx, "Failed to ban user", slog.String("error", err.Error()))
		return err
	}
	logger.InfoContext(ctx, "User banned successfully")
	return nil
}

func (s *adminService) UnbanUser(ctx context.Context, userID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "AuthService"),
		slog.String("operation", "UnbanUser"),
		slog.Int("target_user_id", userID),
	)

	logger.DebugContext(ctx, "Unbanning user")
	if err := s.userRepo.Unban(ctx, userID); err != nil {
		logger.ErrorContext(ctx, "Failed to unban user", slog.String("error", err.Error()))
		return err
	}
	logger.InfoContext(ctx, "User unbanned successfully")
	return nil
}
