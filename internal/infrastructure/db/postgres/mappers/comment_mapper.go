package mappers

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainComment(model *models.Comment) *domain.Comment {
	if model == nil {
		return nil
	}
	return &domain.Comment{
		ID:        model.ID,
		UserID:    model.UserID,
		VideoID:   model.VideoID,
		Content:   model.Content,
		CreatedAt: model.CreatedAt,
	}
}

func FromDomainComment(domainComment *domain.Comment) *models.Comment {
	if domainComment == nil {
		return nil
	}
	return &models.Comment{
		ID:        domainComment.ID,
		UserID:    domainComment.UserID,
		VideoID:   domainComment.VideoID,
		Content:   domainComment.Content,
		CreatedAt: domainComment.CreatedAt,
	}
}

func ToDomainCommentList(models []models.Comment) []*domain.Comment {
	comments := make([]*domain.Comment, len(models))
	for i, m := range models {
		comments[i] = ToDomainComment(&m)
	}
	return comments
}
