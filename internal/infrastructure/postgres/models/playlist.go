package models

import "time"

type Playlist struct {
	ID          int    `gorm:"type:serial"`
	ChannelID   int    `gorm:"not null"`
	Name        string `gorm:"type:varchar(32);not null"`
	Description string
	CreatedAt   time.Time `gorm:"not null;default:current_timestamp"`

	PlaylistVideos []PlaylistsVideo `gorm:"foreignkey:PlaylistID"`
}
