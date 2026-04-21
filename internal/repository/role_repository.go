package repository

import (
	"ZVideo/internal/domain"
	"context"
)

type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	GetByID(ctx context.Context, id int) (*domain.Role, error)
	GetByName(ctx context.Context, name string) (*domain.Role, error)
	GetDefaultRole(ctx context.Context) (*domain.Role, error)
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, id int) error
}
