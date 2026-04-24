package dto

import "time"

type CreatePlaylistRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdatePlaylistRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type PlaylistItemResponse struct {
	VideoID    int    `json:"video_id"`
	VideoTitle string `json:"video_title"`
	Number     int    `json:"number"`
}

type PlaylistResponse struct {
	ID          int                    `json:"id"`
	ChannelID   int                    `json:"channel_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	CreatedAt   time.Time              `json:"created_at"`
	Items       []PlaylistItemResponse `json:"items"`
}
