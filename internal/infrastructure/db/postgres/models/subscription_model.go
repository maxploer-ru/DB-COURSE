package models

import "time"

type Subscription struct {
	UserID         int       `gorm:"not null;primary_key"`
	ChannelID      int       `gorm:"not null;primary_key"`
	NewVideosCount int       `gorm:"not null;default:0"`
	SubscribedAt   time.Time `gorm:"not null;default:current_timestamp"`

	User    User    `gorm:"foreignkey:UserID"`
	Channel Channel `gorm:"foreignkey:ChannelID"`
}
