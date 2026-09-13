package models

import "time"

type Video struct {
	ID               int    `gorm:"type:serial;primaryKey"`
	ChannelID        int    `gorm:"not null"`
	Title            string `gorm:"type:varchar(64);not null"`
	Description      string
	Filepath         string    `gorm:"not null"`
	OriginalFilename string    `gorm:"not null;default:''"`
	Status           string    `gorm:"type:varchar(16);not null;default:'pending'"`
	CreatedAt        time.Time `gorm:"not null;default:current_timestamp"`

	Channel Channel `gorm:"foreignKey:ChannelID"`

	WatchHistory []Viewing `gorm:"foreignKey:VideoID"`
	Comments     []Comment `gorm:"foreignKey:VideoID"`

	PlaylistVideos   []PlaylistItem `gorm:"foreignKey:VideoID"`
	UserVideoRatings []VideoRating  `gorm:"foreignKey:VideoID"`
}
