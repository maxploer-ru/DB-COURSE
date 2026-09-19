package integration_test

import (
	"ZVideo/internal/domain"
	pgrepository "ZVideo/internal/infrastructure/db/postgres/repository"
	"ZVideo/internal/service"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestChannelService_Integration_StateTransition_EquivalencePartitioning(t *testing.T) {
	tx := integrationDB.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { tx.Rollback() })

	ctx := context.Background()
	roleRepo := pgrepository.NewRoleRepository(tx)
	userRepo := pgrepository.NewUserRepository(tx)
	channelRepo := pgrepository.NewChannelRepository(tx)
	role, err := roleRepo.GetDefaultRole(ctx)
	require.NoError(t, err)
	require.NotNil(t, role)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	user := &domain.User{
		Username:             "service_user_" + suffix,
		Email:                "service_" + suffix + "@example.com",
		PasswordHash:         "integration-hash",
		IsActive:             true,
		NotificationsEnabled: true,
		Role:                 role,
	}
	require.NoError(t, userRepo.Create(ctx, user))

	channelService := service.NewChannelService(channelRepo)
	created, err := channelService.CreateChannel(ctx, user.ID, "svc_"+suffix, "description")
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotZero(t, created.ID)

	loaded, err := channelService.GetChannel(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.Name, loaded.Name)

	_, err = channelService.CreateChannel(ctx, user.ID, "dup_"+suffix, "duplicate owner")
	require.ErrorIs(t, err, domain.ErrChannelAlreadyExists)
}
