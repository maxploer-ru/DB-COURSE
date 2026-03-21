package repository

import (
	"ZVideo/internal/domain/entity"
	"context"
)

type RoleRepository interface {
	GetByID(ctx context.Context, id int) (*entity.Role, error)
	GetByName(ctx context.Context, name string) (*entity.Role, error)
	GetDefaultRole(ctx context.Context) (*entity.Role, error)
}
