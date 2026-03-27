package entity

import "time"

type Video struct {
	ID          int
	ChannelID   int
	Title       string
	Description string
	Filepath    string
	CreatedAt   time.Time
}
