package models

import (
	_ "gorm.io/gorm"
	"time"
)

type User struct {
	ID                   int       `gorm:"type:serial"`
	RoleID               int       `gorm:"not null"`
	Username             string    `gorm:"type:varchar(32);unique;not null"`
	Email                string    `gorm:"type:varchar(64);unique;not null"`
	PasswordHash         string    `gorm:"not null"`
	CreatedAt            time.Time `gorm:"not null;default:current_timestamp"`
	NotificationsEnabled bool      `gorm:"not null;default:true"`

	Channels     []Channel `gorm:"foreignkey:UserID"`
	WatchHistory []Viewing `gorm:"foreignkey:UserID"`
	Comments     []Comment `gorm:"foreignkey:UserID"`

	UserChannelSubscriptions []UsersChannelsSubscription `gorm:"foreignkey:UserID"`
	UserVideoRatings         []UsersVideosRating         `gorm:"foreignkey:UserID"`
	UserCommentRatings       []UsersCommentsRating       `gorm:"foreignkey:UserID"`
}
