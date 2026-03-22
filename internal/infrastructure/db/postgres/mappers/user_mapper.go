package mappers

import (
	entity2 "ZVideo/internal/domain/auth/entity"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainUser(model *models.User) *entity2.User {
	if model == nil {
		return nil
	}

	user := &entity2.User{
		ID:        model.ID,
		Username:  model.Username,
		Email:     model.Email,
		RoleID:    model.RoleID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
	user.Role = &entity2.Role{
		ID:        model.Role.ID,
		Name:      model.Role.Name,
		IsDefault: model.Role.IsDefault,
	}
	return user
}

func ToDomainUserList(models []*models.User) []*entity2.User {
	users := make([]*entity2.User, len(models))
	for i, model := range models {
		users[i] = ToDomainUser(model)
	}
	return users
}

func FromDomainUser(user *entity2.User) *models.User {
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

func FromDomainUserList(users []*entity2.User) []*models.User {
	usersModels := make([]*models.User, len(users))
	for i, user := range users {
		usersModels[i] = FromDomainUser(user)
	}
	return usersModels
}
