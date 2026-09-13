package models

import "time"

type User struct {
	ID                   int       `gorm:"type:serial;primaryKey"`
	RoleID               int       `gorm:"not null"`
	Username             string    `gorm:"type:varchar(32);unique;not null"`
	Email                string    `gorm:"type:varchar(64);unique;not null"`
	PasswordHash         string    `gorm:"not null"`
	IsActive             bool      `gorm:"not null;default:true"`
	NotificationsEnabled bool      `gorm:"not null;default:true"`
	CreatedAt            time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt            time.Time `gorm:"not null;autoUpdateTime"`

	Role Role `gorm:"foreignKey:RoleID"`

	Channels     []Channel `gorm:"foreignKey:UserID"`
	WatchHistory []Viewing `gorm:"foreignKey:UserID"`
	Comments     []Comment `gorm:"foreignKey:UserID"`

	UserChannelSubscriptions []Subscription  `gorm:"foreignKey:UserID"`
	UserVideoRatings         []VideoRating   `gorm:"foreignKey:UserID"`
	UserCommentRatings       []CommentRating `gorm:"foreignKey:UserID"`
}
