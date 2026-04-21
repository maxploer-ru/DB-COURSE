package domain

import "time"

type Video struct {
	ID          int
	ChannelID   int
	Title       string
	Description string
	Filepath    string
	CreatedAt   time.Time
}

type VideoStats struct {
	Views    int
	Likes    int
	Dislikes int
	Comments int
}
