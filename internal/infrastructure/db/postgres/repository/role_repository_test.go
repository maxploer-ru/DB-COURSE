package repository

import (
	"ZVideo/internal/domain"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoleRepository_Create(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewRoleRepository(db)

	role := &domain.Role{Name: "role_a", IsDefault: false}
	err := repo.Create(context.Background(), role)
	require.NoError(t, err)
	require.NotZero(t, role.ID)
}

func TestRoleRepository_GetByID(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewRoleRepository(db)

	id := insertRole(t, "role_b", false)
	role, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, id, role.ID)
}

func TestRoleRepository_GetByName(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewRoleRepository(db)

	_ = insertRole(t, "role_c", false)
	role, err := repo.GetByName(context.Background(), "role_c")
	require.NoError(t, err)
	require.Equal(t, "role_c", role.Name)
}

func TestRoleRepository_GetDefaultRole(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewRoleRepository(db)

	role, err := repo.GetDefaultRole(context.Background())
	require.NoError(t, err)
	require.NotNil(t, role)
	require.True(t, role.IsDefault)
}

func TestRoleRepository_Update(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewRoleRepository(db)

	role := &domain.Role{Name: "role_d", IsDefault: false}
	require.NoError(t, repo.Create(context.Background(), role))
	role.Name = "role_d_updated"

	err := repo.Update(context.Background(), role)
	require.NoError(t, err)

	updated, err := repo.GetByID(context.Background(), role.ID)
	require.NoError(t, err)
	require.Equal(t, "role_d_updated", updated.Name)
}

func TestRoleRepository_Delete(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewRoleRepository(db)

	role := &domain.Role{Name: "role_e", IsDefault: false}
	require.NoError(t, repo.Create(context.Background(), role))

	err := repo.Delete(context.Background(), role.ID)
	require.NoError(t, err)

	deleted, err := repo.GetByID(context.Background(), role.ID)
	require.NoError(t, err)
	require.Nil(t, deleted)
}
