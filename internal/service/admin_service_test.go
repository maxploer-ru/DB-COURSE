package service_test

import (
	service "ZVideo/internal/service"
	"ZVideo/mocks"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdminService_BanUser(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	userRepo.On("Ban", ctx, 10).Return(nil)

	svc := service.NewAdminService(userRepo)
	err := svc.BanUser(ctx, 10)
	require.NoError(t, err)
}

func TestAdminService_UnbanUser(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	userRepo.On("Unban", ctx, 11).Return(nil)

	svc := service.NewAdminService(userRepo)
	err := svc.UnbanUser(ctx, 11)
	require.NoError(t, err)
}
