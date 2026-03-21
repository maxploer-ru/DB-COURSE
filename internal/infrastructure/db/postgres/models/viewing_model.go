package models

import "time"

type Viewing struct {
	ID        int       `gorm:"type:serial;primary_key"`
	UserID    int       `gorm:"not null"`
	VideoID   int       `gorm:"not null"`
	WatchedAt time.Time `gorm:"not null;default:current_timestamp"`
}
