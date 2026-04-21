package repository

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/mappers"
	"ZVideo/internal/infrastructure/db/postgres/models"
	"context"

	"gorm.io/gorm"
)

type ViewingRepository struct {
	db *gorm.DB
}

func NewViewingRepository(db *gorm.DB) *ViewingRepository {
	return &ViewingRepository{db: db}
}

func (r *ViewingRepository) Create(ctx context.Context, viewing *domain.Viewing) error {
	model := mappers.FromDomainViewing(viewing)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *ViewingRepository) GetTotalViews(ctx context.Context, videoID int) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Viewing{}).
		Where("video_id = ?", videoID).
		Count(&count).Error
	return int(count), err
}
