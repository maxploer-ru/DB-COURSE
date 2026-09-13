package repository

import (
	"ZVideo/internal/domain"
	"context"
)

type RoleRepository interface {
	GetByID(ctx context.Context, id int) (*domain.Role, error)
	GetByName(ctx context.Context, name string) (*domain.Role, error)
	GetDefaultRole(ctx context.Context) (*domain.Role, error)
}
