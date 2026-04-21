package mappers

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainVideoRating(model *models.VideoRating) *domain.VideoRating {
	if model == nil {
		return nil
	}
	return &domain.VideoRating{
		UserID:  model.UserID,
		VideoID: model.VideoID,
		Liked:   model.Liked,
		RatedAt: model.RatedAt,
	}
}

func FromDomainVideoRating(rating *domain.VideoRating) *models.VideoRating {
	if rating == nil {
		return nil
	}
	return &models.VideoRating{
		UserID:  rating.UserID,
		VideoID: rating.VideoID,
		Liked:   rating.Liked,
		RatedAt: rating.RatedAt,
	}
}
