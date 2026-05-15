package repository

import (
	"ZVideo/internal/domain"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewUserRepository(db)

	roleID := insertRole(t, "role_user", false)
	user := &domain.User{Role: &domain.Role{ID: roleID}, Username: "user_a", Email: "user_a@example.com", PasswordHash: "hash", IsActive: true, NotificationsEnabled: true}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)
	require.NotZero(t, user.ID)
}

func TestUserRepository_GetByID(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewUserRepository(db)

	roleID := insertRole(t, "role_b", false)
	userID := insertUser(t, roleID, "user_b")

	user, err := repo.GetByID(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, userID, user.ID)
	require.NotNil(t, user.Role)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewUserRepository(db)

	roleID := insertRole(t, "role_c", false)
	userID := insertUser(t, roleID, "user_c")

	user, err := repo.GetByEmail(context.Background(), "user_c@example.com")
	require.NoError(t, err)
	require.Equal(t, userID, user.ID)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewUserRepository(db)

	roleID := insertRole(t, "role_d", false)
	userID := insertUser(t, roleID, "user_d")

	user, err := repo.GetByUsername(context.Background(), "user_d")
	require.NoError(t, err)
	require.Equal(t, userID, user.ID)
}

func TestUserRepository_Update(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewUserRepository(db)

	roleID := insertRole(t, "role_e", false)
	userID := insertUser(t, roleID, "user_e")

	user, err := repo.GetByID(context.Background(), userID)
	require.NoError(t, err)
	user.Username = "user_e_new"

	err = repo.Update(context.Background(), user)
	require.NoError(t, err)

	updated, err := repo.GetByID(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, "user_e_new", updated.Username)
}

func TestUserRepository_Delete(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewUserRepository(db)

	roleID := insertRole(t, "role_f", false)
	userID := insertUser(t, roleID, "user_f")

	err := repo.Delete(context.Background(), userID)
	require.NoError(t, err)

	user, err := repo.GetByID(context.Background(), userID)
	require.NoError(t, err)
	require.False(t, user.IsActive)
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewUserRepository(db)

	roleID := insertRole(t, "role_g", false)
	_ = insertUser(t, roleID, "user_g")

	exists, err := repo.ExistsByEmail(context.Background(), "user_g@example.com")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestUserRepository_ExistsByUsername(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewUserRepository(db)

	roleID := insertRole(t, "role_h", false)
	_ = insertUser(t, roleID, "user_h")

	exists, err := repo.ExistsByUsername(context.Background(), "user_h")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestUserRepository_Ban(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewUserRepository(db)

	roleID := insertRole(t, "role_i", false)
	userID := insertUser(t, roleID, "user_i")

	err := repo.Ban(context.Background(), userID)
	require.NoError(t, err)

	user, err := repo.GetByID(context.Background(), userID)
	require.NoError(t, err)
	require.False(t, user.IsActive)
}

func TestUserRepository_Unban(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewUserRepository(db)

	roleID := insertRole(t, "role_j", false)
	userID := insertUser(t, roleID, "user_j")

	_ = repo.Ban(context.Background(), userID)
	err := repo.Unban(context.Background(), userID)
	require.NoError(t, err)

	user, err := repo.GetByID(context.Background(), userID)
	require.NoError(t, err)
	require.True(t, user.IsActive)
}

func TestUserRepository_SetNotificationsEnabled(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewUserRepository(db)

	roleID := insertRole(t, "role_k", false)
	userID := insertUser(t, roleID, "user_k")

	err := repo.SetNotificationsEnabled(context.Background(), userID, false)
	require.NoError(t, err)

	user, err := repo.GetByID(context.Background(), userID)
	require.NoError(t, err)
	require.False(t, user.NotificationsEnabled)
}
