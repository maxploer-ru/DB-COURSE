package models

import "time"

type Video struct {
	ID          int    `gorm:"type:serial;primary_key"`
	ChannelID   int    `gorm:"not null"`
	Title       string `gorm:"type:varchar(64);not null"`
	Description string
	Filepath    string    `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null;default:current_timestamp"`

	Channel Channel `gorm:"foreignkey:ChannelID"`

	WatchHistory []Viewing `gorm:"foreignkey:VideoID"`
	Comments     []Comment `gorm:"foreignkey:VideoID"`

	PlaylistVideos   []PlaylistItem `gorm:"foreignkey:VideoID"`
	UserVideoRatings []VideoRating  `gorm:"foreignkey:VideoID"`
}
