package mappers

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func FromDomainViewing(model *domain.Viewing) *models.Viewing {
	if model == nil {
		return nil
	}
	return &models.Viewing{
		ID:        model.ID,
		UserID:    model.UserID,
		VideoID:   model.VideoID,
		WatchedAt: model.WatchedAt,
	}
}
