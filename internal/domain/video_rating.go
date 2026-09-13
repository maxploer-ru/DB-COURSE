package domain

import "time"

type RatingAction string

const (
	RatingActionLike    RatingAction = "like"
	RatingActionDislike RatingAction = "dislike"
	RatingActionRemove  RatingAction = "remove"
)

type VideoRating struct {
	UserID  int
	VideoID int
	Liked   bool
	RatedAt time.Time
}
