package domain

import "time"

const (
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
	RoleUser      = "user"
)

type User struct {
	ID                   int
	Username             string
	Email                string
	PasswordHash         string
	IsActive             bool
	NotificationsEnabled bool
	CreatedAt            time.Time
	UpdatedAt            time.Time

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
