package mappers

import (
	"ZVideo/internal/domain/entity"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainUser(model *models.User) *entity.User {
	if model == nil {
		return nil
	}

	user := &entity.User{
		ID:        model.ID,
		Username:  model.Username,
		Email:     model.Email,
		RoleID:    model.RoleID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
	user.Role = &entity.Role{
		ID:        model.Role.ID,
		Name:      model.Role.Name,
		IsDefault: model.Role.IsDefault,
	}
	return user
}

func ToDomainUserList(models []*models.User) []*entity.User {
	users := make([]*entity.User, len(models))
	for i, model := range models {
		users[i] = ToDomainUser(model)
	}
	return users
}

func FromDomainUser(user *entity.User) *models.User {
	if user == nil {
		return nil
	}

	return &models.User{
		ID:                   user.ID,
		RoleID:               user.RoleID,
		Username:             user.Username,
		Email:                user.Email,
		PasswordHash:         user.PasswordHash,
		NotificationsEnabled: user.NotificationsEnabled,
		IsActive:             user.IsActive,
		CreatedAt:            user.CreatedAt,
		UpdatedAt:            user.UpdatedAt,
	}
}

func FromDomainUserList(users []*entity.User) []*models.User {
	usersModels := make([]*models.User, len(users))
	for i, user := range users {
		usersModels[i] = FromDomainUser(user)
	}
	return usersModels
}
