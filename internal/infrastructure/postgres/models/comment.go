package models

import "time"

type Comment struct {
	ID        int       `gorm:"type:serial"`
	UserID    int       `gorm:"not null"`
	VideoID   int       `gorm:"not null"`
	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp"`

	UserCommentRatings []UsersCommentsRating `gorm:"foreignkey:CommentID"`
}
