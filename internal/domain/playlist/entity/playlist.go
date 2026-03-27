package entity

import "time"

type Playlist struct {
	ID          int
	ChannelID   int
	Name        string
	Description string
	CreatedAt   time.Time
}
