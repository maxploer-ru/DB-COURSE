package mappers

import (
	"ZVideo/internal/delivery/handlers/dto"
	"ZVideo/internal/domain"
)

func ToGetChannelResponse(channel *domain.Channel, subscribersCount int) *dto.GetChannelResponse {
	return &dto.GetChannelResponse{
		ID:               channel.ID,
		UserID:           channel.UserID,
		Name:             channel.Name,
		Description:      channel.Description,
		SubscribersCount: subscribersCount,
	}
}
