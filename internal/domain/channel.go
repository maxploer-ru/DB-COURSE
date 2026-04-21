package domain

import "time"

type Channel struct {
	ID          int
	UserID      int
	Name        string
	Description string
	CreatedAt   time.Time
}
