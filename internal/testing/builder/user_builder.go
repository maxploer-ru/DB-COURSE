package builder

import (
	"ZVideo/internal/domain"
	"time"
)

type UserBuilder struct {
	user *domain.User
}

func NewUserBuilder() *UserBuilder {
	return &UserBuilder{
		user: &domain.User{
			ID:                   1,
			Username:             "test_user",
			Email:                "test@example.com",
			PasswordHash:         "hashed_pwd",
			IsActive:             true,
			NotificationsEnabled: true,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
			Role:                 &domain.Role{ID: 1, Name: domain.RoleUser, IsDefault: true},
		},
	}
}

func (b *UserBuilder) WithID(id int) *UserBuilder {
	b.user.ID = id
	return b
}

func (b *UserBuilder) Banned() *UserBuilder {
	b.user.IsActive = false
	return b
}

func (b *UserBuilder) Build() *domain.User {
	return b.user
}
