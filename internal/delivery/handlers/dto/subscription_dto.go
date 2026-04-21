package dto

import "time"

type SubscriptionResponse struct {
	ChannelID      int       `json:"channel_id"`
	ChannelName    string    `json:"channel_name"`
	NewVideosCount int       `json:"new_videos_count"`
	SubscribedAt   time.Time `json:"subscribed_at"`
}

type SubscriptionChannelResponse struct {
	ID               int       `json:"id"`
	UserID           int       `json:"user_id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	SubscribersCount int       `json:"subscribers_count"`
	NewVideosCount   int       `json:"new_videos_count"`
	SubscribedAt     time.Time `json:"subscribed_at"`
}
