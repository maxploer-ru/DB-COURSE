package dto

type CreateChannelRequest struct {
	ChannelName string `json:"channel_name"`
	Description string `json:"description"`
}

type UpdateChannelRequest struct {
	ChannelName *string `json:"channel_name"`
	Description *string `json:"description"`
}

type GetChannelResponse struct {
	ID               int    `json:"id"`
	UserID           int    `json:"user_id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	SubscribersCount int    `json:"subscribers_count"`
}
