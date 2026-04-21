package repository

import (
	"ZVideo/internal/domain"
	"context"
)

type ChannelRepository interface {
	Create(ctx context.Context, channel *domain.Channel) error
	GetByID(ctx context.Context, id int) (*domain.Channel, error)
	GetByUserID(ctx context.Context, userID int) (*domain.Channel, error)
	GetByName(ctx context.Context, name string) (*domain.Channel, error)
	Update(ctx context.Context, channel *domain.Channel) error
	Delete(ctx context.Context, id int) error
	ExistsByName(ctx context.Context, name string) (bool, error)
}
