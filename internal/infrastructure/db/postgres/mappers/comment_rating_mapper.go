package mappers

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainCommentRating(model *models.CommentRating) *domain.CommentRating {
	if model == nil {
		return nil
	}
	return &domain.CommentRating{
		UserID:    model.UserID,
		CommentID: model.CommentID,
		Liked:     model.Liked,
		RatedAt:   model.RatedAt,
	}
}

func FromDomainCommentRating(rating *domain.CommentRating) *models.CommentRating {
	if rating == nil {
		return nil
	}
	return &models.CommentRating{
		UserID:    rating.UserID,
		CommentID: rating.CommentID,
		Liked:     rating.Liked,
		RatedAt:   rating.RatedAt,
	}
}
