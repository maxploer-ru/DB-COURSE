package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"fmt"
	"log/slog"
	"strings"
)

type AdminService interface {
	BanUser(ctx context.Context, adminID, targetUserID int) error
	UnbanUser(ctx context.Context, adminID, targetUserID int) error
	ChangeUserRole(ctx context.Context, adminID, targetUserID int, roleName string) error
	ListUsers(ctx context.Context, adminID, limit, offset int) (*domain.PageResponse[*domain.User], error)
}

type adminService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

func NewAdminService(userRepo repository.UserRepository, roleRepo repository.RoleRepository) AdminService {
	return &adminService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

func (s *adminService) BanUser(ctx context.Context, adminID, targetUserID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "AdminService"),
		slog.String("operation", "BanUser"),
		slog.Int("admin_id", adminID),
		slog.Int("target_user_id", targetUserID),
	)

	if err := s.userRepo.Ban(ctx, targetUserID); err != nil {
		logger.ErrorContext(ctx, "Failed to ban user", slog.String("error", err.Error()))
		return fmt.Errorf("ban user failed: %w", err)
	}
	logger.InfoContext(ctx, "User banned successfully")
	return nil
}

func (s *adminService) UnbanUser(ctx context.Context, adminID, targetUserID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "AdminService"),
		slog.String("operation", "UnbanUser"),
		slog.Int("admin_id", adminID),
		slog.Int("target_user_id", targetUserID),
	)

	if err := s.userRepo.Unban(ctx, targetUserID); err != nil {
		logger.ErrorContext(ctx, "Failed to unban user", slog.String("error", err.Error()))
		return fmt.Errorf("unban user failed: %w", err)
	}
	logger.InfoContext(ctx, "User unbanned successfully")
	return nil
}

func (s *adminService) ChangeUserRole(ctx context.Context, adminID, targetUserID int, roleName string) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "AdminService"),
		slog.String("operation", "ChangeUserRole"),
		slog.Int("admin_id", adminID),
		slog.Int("target_user_id", targetUserID),
		slog.String("target_role", roleName),
	)

	if adminID == targetUserID {
		logger.WarnContext(ctx, "Admin attempted to change their own role")
		return domain.ErrForbidden
	}

	roleName = strings.TrimSpace(strings.ToLower(roleName))
	role, err := s.roleRepo.GetByName(ctx, roleName)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to find role", slog.String("error", err.Error()))
		return fmt.Errorf("find role failed: %w", err)
	}
	if role == nil {
		return domain.ErrRoleNotFound
	}

	user, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get user", slog.String("error", err.Error()))
		return fmt.Errorf("get user failed: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	if user.Role != nil && strings.EqualFold(user.Role.Name, role.Name) {
		return nil
	}

	user.Role = role
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.ErrorContext(ctx, "Failed to update user role", slog.String("error", err.Error()))
		return fmt.Errorf("update user role failed: %w", err)
	}

	logger.InfoContext(ctx, "User role changed successfully")
	return nil
}

func (s *adminService) ListUsers(ctx context.Context, adminID, limit, offset int) (*domain.PageResponse[*domain.User], error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "AdminService"),
		slog.String("operation", "ListUsers"),
		slog.Int("admin_id", adminID),
	)

	logger.DebugContext(ctx, "Listing users")
	users, err := s.userRepo.ListUsers(ctx, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list users", slog.String("error", err.Error()))
		return nil, fmt.Errorf("list users failed: %w", err)
	}
	total, err := s.userRepo.CountUsers(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to count users", slog.String("error", err.Error()))
		return nil, fmt.Errorf("count users failed: %w", err)
	}

	logger.DebugContext(ctx, "Users retrieved successfully", slog.Int("count", len(users)))
	return &domain.PageResponse[*domain.User]{
		Items:      users,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}
