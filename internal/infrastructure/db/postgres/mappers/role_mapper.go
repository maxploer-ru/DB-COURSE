package mappers

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainRole(model *models.Role) *domain.Role {
	if model == nil {
		return nil
	}
	return &domain.Role{
		ID:        model.ID,
		Name:      model.Name,
		IsDefault: model.IsDefault,
	}
}
func FromDomainRole(model *domain.Role) *models.Role {
	if model == nil {
		return nil
	}
	return &models.Role{
		ID:        model.ID,
		Name:      model.Name,
		IsDefault: model.IsDefault,
	}
}
