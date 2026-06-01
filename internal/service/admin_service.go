package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"log/slog"
	"strings"
)

type AdminService interface {
	BanUser(ctx context.Context, userID int) error
	UnbanUser(ctx context.Context, userID int) error
	ChangeUserRole(ctx context.Context, adminID, targetUserID int, roleName string) error
}

type adminService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

func NewAdminService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
) AdminService {
	return &adminService{
		userRepo: userRepo,
		roleRepo: roleRepo,
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

func (s *adminService) ChangeUserRole(ctx context.Context, adminID, targetUserID int, roleName string) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "AuthService"),
		slog.String("operation", "ChangeUserRole"),
		slog.Int("admin_id", adminID),
		slog.Int("target_user_id", targetUserID),
		slog.String("role", roleName),
	)

	if adminID == targetUserID {
		logger.WarnContext(ctx, "Admin cannot change their own role")
		return domain.ErrForbidden
	}

	roleName = strings.TrimSpace(strings.ToLower(roleName))
	if roleName == "" {
		logger.WarnContext(ctx, "Role name is empty")
		return domain.ErrRoleNotFound
	}

	role, err := s.roleRepo.GetByName(ctx, roleName)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get role by name", slog.String("error", err.Error()))
		return err
	}
	if role == nil {
		logger.WarnContext(ctx, "Role not found")
		return domain.ErrRoleNotFound
	}

	user, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get user by id", slog.String("error", err.Error()))
		return err
	}
	if user == nil {
		logger.WarnContext(ctx, "User not found")
		return domain.ErrUserNotFound
	}

	if user.Role != nil && strings.EqualFold(user.Role.Name, role.Name) {
		logger.DebugContext(ctx, "User already has requested role")
		return nil
	}

	user.Role = role
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.ErrorContext(ctx, "Failed to update user role", slog.String("error", err.Error()))
		return err
	}

	logger.InfoContext(ctx, "User role changed successfully")
	return nil
}
