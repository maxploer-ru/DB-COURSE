package domain

import "time"

type Playlist struct {
	ID          int
	ChannelID   int
	Name        string
	Description string
	CreatedAt   time.Time
}

type PlaylistItem struct {
	PlaylistID int
	VideoID    int
	Number     int
	AddedAt    time.Time

	VideoTitle  string
	ChannelName string
	VideoStatus VideoStatus
}
