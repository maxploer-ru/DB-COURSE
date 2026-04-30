package models

import "time"

type User struct {
	ID                   int    `bson:"_id"`
	Username             string `bson:"username"`
	Email                string `bson:"email"`
	PasswordHash         string `bson:"password_hash"`
	IsActive             bool   `bson:"is_active"`
	NotificationsEnabled bool   `bson:"notifications_enabled"`
	RoleID               int    `bson:"role_id"`
}

type Role struct {
	ID        int    `bson:"_id"`
	Name      string `bson:"name"`
	IsDefault bool   `bson:"is_default"`
}

type Channel struct {
	ID          int       `bson:"_id"`
	UserID      int       `bson:"user_id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	CreatedAt   time.Time `bson:"created_at"`
}

type Video struct {
	ID          int       `bson:"_id"`
	ChannelID   int       `bson:"channel_id"`
	Title       string    `bson:"title"`
	Description string    `bson:"description"`
	Filepath    string    `bson:"filepath"`
	CreatedAt   time.Time `bson:"created_at"`
}

type Comment struct {
	ID        int       `bson:"_id"`
	UserID    int       `bson:"user_id"`
	VideoID   int       `bson:"video_id"`
	Content   string    `bson:"content"`
	CreatedAt time.Time `bson:"created_at"`
}

type CommentRating struct {
	ID        string    `bson:"_id"`
	UserID    int       `bson:"user_id"`
	CommentID int       `bson:"comment_id"`
	Liked     bool      `bson:"liked"`
	RatedAt   time.Time `bson:"rated_at"`
}

type VideoRating struct {
	ID      string    `bson:"_id"`
	UserID  int       `bson:"user_id"`
	VideoID int       `bson:"video_id"`
	Liked   bool      `bson:"liked"`
	RatedAt time.Time `bson:"rated_at"`
}

type Subscription struct {
	ID             string    `bson:"_id"`
	UserID         int       `bson:"user_id"`
	ChannelID      int       `bson:"channel_id"`
	NewVideosCount int       `bson:"new_videos_count"`
	SubscribedAt   time.Time `bson:"subscribed_at"`
}

type Viewing struct {
	ID        int       `bson:"_id"`
	UserID    int       `bson:"user_id"`
	VideoID   int       `bson:"video_id"`
	WatchedAt time.Time `bson:"watched_at"`
}

type CommunityPost struct {
	ID        int       `bson:"_id"`
	ChannelID int       `bson:"channel_id"`
	UserID    int       `bson:"user_id"`
	Content   string    `bson:"content"`
	CreatedAt time.Time `bson:"created_at"`
}

type CommunityComment struct {
	ID        int       `bson:"_id"`
	PostID    int       `bson:"post_id"`
	UserID    int       `bson:"user_id"`
	Content   string    `bson:"content"`
	CreatedAt time.Time `bson:"created_at"`
}

type Playlist struct {
	ID          int            `bson:"_id"`
	ChannelID   int            `bson:"channel_id"`
	Name        string         `bson:"name"`
	Description string         `bson:"description"`
	CreatedAt   time.Time      `bson:"created_at"`
	Items       []PlaylistItem `bson:"items"`
}

type PlaylistItem struct {
	VideoID int `bson:"video_id"`
	Number  int `bson:"number"`
}
