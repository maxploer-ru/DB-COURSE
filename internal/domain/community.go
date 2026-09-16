package domain

import "time"

type CommunityPost struct {
	ID        int
	ChannelID int
	UserID    int
	Username  string
	Content   string
	CreatedAt time.Time
}

type CommunityComment struct {
	ID        int
	PostID    int
	UserID    int
	Username  string
	Content   string
	CreatedAt time.Time
}

type Community struct {
	Channel    *Channel
	Posts      []*CommunityPost
	TotalCount int64
	Limit      int
	Offset     int
}
