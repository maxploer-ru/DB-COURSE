package domain

import "time"

type Comment struct {
	ID        int
	UserID    int
	VideoID   int
	Content   string
	CreatedAt time.Time
}
