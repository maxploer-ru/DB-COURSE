package domain

import "time"

type Comment struct {
	ID        int
	UserID    int
	Username  string
	VideoID   int
	Content   string
	CreatedAt time.Time
}
