package repository

import (
	"ZVideo/internal/domain"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	Ban(ctx context.Context, id int) error
	Unban(ctx context.Context, id int) error
	SetNotificationsEnabled(ctx context.Context, id int, enabled bool) error
}
