package repositories

import (
	"ZVideo/internal/domain/video/entity"
	"context"

	"gorm.io/gorm"
)

type ViewingRepository struct {
	db *gorm.DB
}

func NewViewingRepository(db *gorm.DB) *ViewingRepository {
	return &ViewingRepository{
		db: db,
	}
}

func (r *ViewingRepository) Create(ctx context.Context, viewing *entity.Viewing) error {
	return r.db.WithContext(ctx).Create(viewing).Error
}
