package models

import "time"

type PlaylistItem struct {
	PlaylistID int       `gorm:"not null;primaryKey"`
	VideoID    int       `gorm:"not null;primaryKey"`
	Number     int       `gorm:"not null"`
	AddedAt    time.Time `gorm:"not null;default:current_timestamp"`

	Playlist Playlist `gorm:"foreignKey:PlaylistID"`
	Video    Video    `gorm:"foreignKey:VideoID"`
}
