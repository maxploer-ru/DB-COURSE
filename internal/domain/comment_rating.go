package domain

import "time"

type CommentRating struct {
	UserID    int
	CommentID int
	Liked     bool
	RatedAt   time.Time
}
