package dto

import "time"

type CreateCommunityPostRequest struct {
	Content string `json:"content"`
}

type UpdateCommunityPostRequest struct {
	Content string `json:"content"`
}

type CreateCommunityCommentRequest struct {
	Content string `json:"content"`
}

type CommunityCommentResponse struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type CommunityPostResponse struct {
	ID        int                        `json:"id"`
	ChannelID int                        `json:"channel_id"`
	UserID    int                        `json:"user_id"`
	Username  string                     `json:"username"`
	Content   string                     `json:"content"`
	CreatedAt time.Time                  `json:"created_at"`
	Comments  []CommunityCommentResponse `json:"comments"`
}

type CommunityResponse struct {
	Channel GetChannelResponse      `json:"channel"`
	Posts   []CommunityPostResponse `json:"posts"`
}
