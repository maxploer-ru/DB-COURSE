package models

import "time"

type VideoRating struct {
	UserID  int       `gorm:"not null;primaryKey"`
	VideoID int       `gorm:"not null;primaryKey"`
	Liked   bool      `gorm:"not null"`
	RatedAt time.Time `gorm:"not null;default:current_timestamp"`

	User  User  `gorm:"foreignkey:UserID"`
	Video Video `gorm:"foreignkey:VideoID"`
}
