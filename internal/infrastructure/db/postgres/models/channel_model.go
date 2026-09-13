package models

import "time"

type Channel struct {
	ID          int    `gorm:"type:serial;primary_key"`
	UserID      int    `gorm:"not null;unique"`
	Name        string `gorm:"type:varchar(32);unique;not null"`
	Description string
	CreatedAt   time.Time `gorm:"not null;default:current_timestamp"`

	User User `gorm:"foreignKey:UserID"`

	Videos    []Video    `gorm:"foreignKey:ChannelID"`
	Playlists []Playlist `gorm:"foreignKey:ChannelID"`

	UserChannelSubscriptions []Subscription `gorm:"foreignKey:ChannelID"`
}
