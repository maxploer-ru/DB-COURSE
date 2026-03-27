package repository

import (
	"ZVideo/internal/domain/channel/entity"
	"context"
)

type ChannelRepository interface {
	Create(ctx context.Context, channel *entity.Channel) error
	GetByID(ctx context.Context, id int) (*entity.Channel, error)
	GetByUserID(ctx context.Context, userID int) ([]*entity.Channel, error)
	GetByName(ctx context.Context, name string) (*entity.Channel, error)
	Update(ctx context.Context, channel *entity.Channel) error
	Delete(ctx context.Context, id int) error
	ExistsByName(ctx context.Context, name string) (bool, error)
}
