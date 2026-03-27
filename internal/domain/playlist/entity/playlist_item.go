package entity

import "time"

type PlaylistItem struct {
	PlaylistID int
	VideoID    int
	Position   int
	AddedAt    time.Time
}
