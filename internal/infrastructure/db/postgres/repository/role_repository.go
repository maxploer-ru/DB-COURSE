package repository

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/mappers"
	"ZVideo/internal/infrastructure/db/postgres/models"
	"context"
	"errors"
	"fmt"

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

func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	model := mappers.FromDomainRole(role)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("create role failed: %w", err)
	}
	role.ID = model.ID
	return nil
}

func (r *RoleRepository) GetByID(ctx context.Context, id int) (*domain.Role, error) {
	var role domain.Role

	err := r.db.WithContext(ctx).
		First(&role, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get role failed: %w", err)
	}

	return &role, nil
}

func (r *RoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role

	err := r.db.WithContext(ctx).
		First(&role, "name = ?", name).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get role failed: %w", err)
	}

	return &role, nil
}

func (r *RoleRepository) GetDefaultRole(ctx context.Context) (*domain.Role, error) {
	var role domain.Role

	err := r.db.WithContext(ctx).
		First(&role, "is_default = ?", true).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get default role failed: %w", err)
	}

	return &role, nil
}

func (r *RoleRepository) Update(ctx context.Context, role *domain.Role) error {
	model := mappers.FromDomainRole(role)
	err := r.db.WithContext(ctx).Model(&models.Role{}).
		Where("id = ?", model.ID).
		Updates(model).Error
	if err != nil {
		return fmt.Errorf("update role failed: %w", err)
	}
	return nil
}

func (r *RoleRepository) Delete(ctx context.Context, id int) error {
	err := r.db.WithContext(ctx).Delete(&models.Role{}, id).Error
	if err != nil {
		return fmt.Errorf("delete role failed: %w", err)
	}
	return nil
}
