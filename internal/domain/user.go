package domain

import "time"

const (
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
	RoleUser      = "user"
)

type User struct {
	ID                   int       `json:"id"`
	Username             string    `json:"username"`
	Email                string    `json:"email"`
	PasswordHash         string    `json:"-"`
	IsActive             bool      `json:"isActive"`
	NotificationsEnabled bool      `json:"notificationsEnabled"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`

	Role *Role `json:"role,omitempty"`
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
