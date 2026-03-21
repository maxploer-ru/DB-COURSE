package entity

import (
	"errors"
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

func (u *User) Validate() error {
	if u.Username == "" {
		return errors.New("username is required")
	}
	if len(u.Username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	return nil
}
