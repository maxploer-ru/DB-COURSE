package models

import "time"

type User struct {
	ID                   int       `gorm:"type:serial;primary_key'"`
	RoleID               int       `gorm:"not null"`
	Username             string    `gorm:"type:varchar(32);unique;not null"`
	Email                string    `gorm:"type:varchar(64);unique;not null"`
	PasswordHash         string    `gorm:"not null"`
	NotificationsEnabled bool      `gorm:"not null;default:true"`
	IsActive             bool      `gorm:"not null;default:true"`
	CreatedAt            time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt            time.Time `gorm:"not null;autoUpdateTime"`

	Role Role `gorm:"foreignkey:RoleID"`

	Channels     []Channel `gorm:"foreignkey:UserID"`
	WatchHistory []Viewing `gorm:"foreignkey:UserID"`
	Comments     []Comment `gorm:"foreignkey:UserID"`

	UserChannelSubscriptions []Subscription  `gorm:"foreignkey:UserID"`
	UserVideoRatings         []VideoRating   `gorm:"foreignkey:UserID"`
	UserCommentRatings       []CommentRating `gorm:"foreignkey:UserID"`
}
