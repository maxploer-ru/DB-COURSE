package repository

import (
	"ZVideo/internal/domain"
	"context"
)

type ViewingRepository interface {
	Create(ctx context.Context, viewing *domain.Viewing) error
	GetTotalViews(ctx context.Context, videoID int) (int, error)
}
