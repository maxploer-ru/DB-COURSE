package domain

import "time"

type User struct {
	ID                   int
	Username             string
	Email                string
	PasswordHash         string
	IsActive             bool
	NotificationsEnabled bool

	Role *Role
}

type Role struct {
	ID        int
	Name      string
	IsDefault bool
}

type AccessTokenData struct {
	UserID   int
	UserName string
	Role     string
}

type RefreshTokenData struct {
	UserID    int
	TokenID   string
	ExpiresAt time.Time
}
