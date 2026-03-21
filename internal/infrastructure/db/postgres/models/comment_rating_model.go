package models

import "time"

type CommentRating struct {
	UserID    int       `gorm:"not null;primaryKey"`
	CommentID int       `gorm:"not null;primaryKey"`
	Liked     bool      `gorm:"not null"`
	RatedAt   time.Time `gorm:"not null;default:current_timestamp"`

	User    User    `gorm:"foreignkey:UserID"`
	Comment Comment `gorm:"foreignkey:CommentID"`
}
