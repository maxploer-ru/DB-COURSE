package repository

import (
	"ZVideo/internal/domain"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelRepository_Create(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewChannelRepository(db)

	roleID := insertRole(t, "role_a", false)
	userID := insertUser(t, roleID, "user_a")

	ch := &domain.Channel{UserID: userID, Name: "channel_a", Description: "d"}
	err := repo.Create(context.Background(), ch)
	require.NoError(t, err)
	require.NotZero(t, ch.ID)
}

func TestChannelRepository_GetByID(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewChannelRepository(db)

	roleID := insertRole(t, "role_b", false)
	userID := insertUser(t, roleID, "user_b")
	channelID := insertChannel(t, userID, "channel_b")

	ch, err := repo.GetByID(context.Background(), channelID)
	require.NoError(t, err)
	require.Equal(t, channelID, ch.ID)
}

func TestChannelRepository_GetByUserID(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewChannelRepository(db)

	roleID := insertRole(t, "role_c", false)
	userID := insertUser(t, roleID, "user_c")
	channelID := insertChannel(t, userID, "channel_c")

	ch, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, channelID, ch.ID)
}

func TestChannelRepository_GetByName(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewChannelRepository(db)

	roleID := insertRole(t, "role_d", false)
	userID := insertUser(t, roleID, "user_d")
	channelID := insertChannel(t, userID, "channel_d")

	ch, err := repo.GetByName(context.Background(), "channel_d")
	require.NoError(t, err)
	require.Equal(t, channelID, ch.ID)
}

func TestChannelRepository_Update(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewChannelRepository(db)

	roleID := insertRole(t, "role_e", false)
	userID := insertUser(t, roleID, "user_e")
	channelID := insertChannel(t, userID, "channel_e")

	ch, err := repo.GetByID(context.Background(), channelID)
	require.NoError(t, err)
	ch.Description = "new"

	err = repo.Update(context.Background(), ch)
	require.NoError(t, err)

	updated, err := repo.GetByID(context.Background(), channelID)
	require.NoError(t, err)
	require.Equal(t, "new", updated.Description)
}

func TestChannelRepository_Delete(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewChannelRepository(db)

	roleID := insertRole(t, "role_f", false)
	userID := insertUser(t, roleID, "user_f")
	channelID := insertChannel(t, userID, "channel_f")

	err := repo.Delete(context.Background(), channelID)
	require.NoError(t, err)

	ch, err := repo.GetByID(context.Background(), channelID)
	require.NoError(t, err)
	require.Nil(t, ch)
}

func TestChannelRepository_ExistsByName(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewChannelRepository(db)

	roleID := insertRole(t, "role_g", false)
	userID := insertUser(t, roleID, "user_g")
	_ = insertChannel(t, userID, "channel_g")

	exists, err := repo.ExistsByName(context.Background(), "channel_g")
	require.NoError(t, err)
	require.True(t, exists)
}
