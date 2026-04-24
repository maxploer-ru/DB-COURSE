package domain

import "time"

type Playlist struct {
	ID          int
	ChannelID   int
	Name        string
	Description string
	CreatedAt   time.Time
	Items       []PlaylistItem
}

type PlaylistItem struct {
	VideoID    int
	VideoTitle string
	Number     int
}
