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

type CommunityPostWithComments struct {
	Post     *CommunityPost
	Comments []*CommunityComment
}

type Community struct {
	Channel *Channel
	Posts   []*CommunityPostWithComments
}
