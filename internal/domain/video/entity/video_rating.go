package entity

import "time"

type VideoRating struct {
	UserID  int
	VideoID int
	Liked   bool
	RatedAt time.Time
}
