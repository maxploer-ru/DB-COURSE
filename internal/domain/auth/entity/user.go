package entity

import (
	"time"
)

type User struct {
	ID                   int
	RoleID               int
	Username             string
	Email                string
	PasswordHash         string
	NotificationsEnabled bool
	IsActive             bool
	CreatedAt            time.Time
	UpdatedAt            time.Time

	Role *Role
}
