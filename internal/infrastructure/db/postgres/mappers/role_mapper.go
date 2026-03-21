package mappers

import (
	"ZVideo/internal/domain/entity"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainRole(model *models.Role) *entity.Role {
	if model == nil {
		return nil
	}
	return &entity.Role{
		ID:        model.ID,
		Name:      model.Name,
		IsDefault: model.IsDefault,
	}
}
func FromDomainRole(model *entity.Role) *models.Role {
	if model == nil {
		return nil
	}
	return &models.Role{
		ID:        model.ID,
		Name:      model.Name,
		IsDefault: model.IsDefault,
	}
}
