package repository

import (
	"ZVideo/internal/domain/auth/entity"
	"context"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}
