package models

import "time"

type Channel struct {
	ID          int    `gorm:"type:serial"`
	UserID      int    `gorm:"not null"`
	Name        string `gorm:"type:varchar(32);unique;not null"`
	Description string
	CreatedAt   time.Time `gorm:"not null;default:current_timestamp"`

	Videos    []Video    `gorm:"foreignkey:ChannelID"`
	Playlists []Playlist `gorm:"foreignkey:ChannelID"`

	UserChannelSubscriptions []UsersChannelsSubscription `gorm:"foreignkey:ChannelID"`
}
