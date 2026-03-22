package repositories

import (
	"ZVideo/internal/domain/auth/entity"
	"context"

	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{
		db: db,
	}
}

func (repo *RoleRepository) GetByID(ctx context.Context, id int) (*entity.Role, error) {
	var role entity.Role

	err := repo.db.WithContext(ctx).
		First(&role, "id = ?", id).Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (repo *RoleRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	var role entity.Role

	err := repo.db.WithContext(ctx).
		First(&role, "name = ?", name).Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (repo *RoleRepository) GetDefaultRole(ctx context.Context) (*entity.Role, error) {
	var role entity.Role

	err := repo.db.WithContext(ctx).
		First(&role, "is_default = ?", true).Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}
