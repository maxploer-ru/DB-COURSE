package models

import "time"

type CommunityPost struct {
	ID        int       `gorm:"type:serial;primary_key"`
	ChannelID int       `gorm:"not null"`
	UserID    int       `gorm:"not null"`
	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp"`

	User     User               `gorm:"foreignKey:UserID"`
	Comments []CommunityComment `gorm:"foreignKey:PostID"`
}

type CommunityComment struct {
	ID        int       `gorm:"type:serial;primary_key"`
	PostID    int       `gorm:"not null"`
	UserID    int       `gorm:"not null"`
	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp"`

	User User `gorm:"foreignKey:UserID"`
}

func (CommunityComment) TableName() string {
	return "community_post_comments"
}
