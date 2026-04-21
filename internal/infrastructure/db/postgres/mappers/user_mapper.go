package mappers

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainUser(model *models.User) *domain.User {
	if model == nil {
		return nil
	}

	user := &domain.User{
		ID:                   model.ID,
		Username:             model.Username,
		Email:                model.Email,
		PasswordHash:         model.PasswordHash,
		IsActive:             model.IsActive,
		NotificationsEnabled: model.NotificationsEnabled,
	}
	user.Role = &domain.Role{
		ID:        model.Role.ID,
		Name:      model.Role.Name,
		IsDefault: model.Role.IsDefault,
	}
	return user
}

func ToDomainUserList(models []*models.User) []*domain.User {
	users := make([]*domain.User, len(models))
	for i, model := range models {
		users[i] = ToDomainUser(model)
	}
	return users
}

func FromDomainUser(user *domain.User) *models.User {
	if user == nil {
		return nil
	}

	return &models.User{
		ID:                   user.ID,
		RoleID:               user.Role.ID, // TODO: проверять что не nil
		Username:             user.Username,
		Email:                user.Email,
		PasswordHash:         user.PasswordHash,
		IsActive:             user.IsActive,
		NotificationsEnabled: user.NotificationsEnabled,
	}
}

func FromDomainUserList(users []*domain.User) []*models.User {
	usersModels := make([]*models.User, len(users))
	for i, user := range users {
		usersModels[i] = FromDomainUser(user)
	}
	return usersModels
}
