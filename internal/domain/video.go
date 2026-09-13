package domain

import "time"

type VideoStatus string

const (
	VideoStatusPending VideoStatus = "pending"
	VideoStatusReady   VideoStatus = "ready"
)

type Video struct {
	ID               int
	ChannelID        int
	ChannelName      string
	Title            string
	Description      string
	Filepath         string
	OriginalFilename string
	Status           VideoStatus
	CreatedAt        time.Time

	Stats *VideoStats
}

type VideoStats struct {
	Views    int
	Likes    int
	Dislikes int
	Comments int
}
