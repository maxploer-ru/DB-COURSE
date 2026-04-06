package repository

import (
	"ZVideo/internal/domain/auth/entity"
	"context"
)

type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) error
	GetByID(ctx context.Context, id int) (*entity.Role, error)
	GetByName(ctx context.Context, name string) (*entity.Role, error)
	GetDefaultRole(ctx context.Context) (*entity.Role, error)
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id int) error
}
