package domain

import "time"

type Channel struct {
	ID            int
	UserID        int
	OwnerUsername string
	Name          string
	Description   string
	CreatedAt     time.Time
}
