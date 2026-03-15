package models

import "time"

type Video struct {
	ID              int    `gorm:"type:serial"`
	ChannelID       int    `gorm:"not null"`
	Title           string `gorm:"type:varchar(64);not null"`
	Description     string
	Filepath        string    `gorm:"not null"`
	PreviewFilepath string    `gorm:"not null"`
	CreatedAt       time.Time `gorm:"not null;default:current_timestamp"`

	WatchHistory []Viewing `gorm:"foreignkey:VideoID"`
	Comments     []Comment `gorm:"foreignkey:VideoID"`

	PlaylistVideos   []PlaylistsVideo    `gorm:"foreignkey:VideoID"`
	UserVideoRatings []UsersVideosRating `gorm:"foreignkey:VideoID"`
}
