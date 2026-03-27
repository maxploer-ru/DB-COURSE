package repository

import (
	"ZVideo/internal/domain/video/entity"
	"context"
)

type ViewingRepository interface {
	Create(ctx context.Context, viewing *entity.Viewing) error
}
