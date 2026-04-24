package mappers

import (
	"ZVideo/internal/delivery/handlers/dto"
	"ZVideo/internal/domain"
)

func ToSubscriptionResponse(sub *domain.Subscription, channelName string) *dto.SubscriptionResponse {
	if sub == nil {
		return nil
	}
	return &dto.SubscriptionResponse{
		ChannelID:      sub.ChannelID,
		ChannelName:    channelName,
		NewVideosCount: sub.NewVideosCount,
		SubscribedAt:   sub.SubscribedAt,
	}
}

func ToSubscriptionChannelResponse(sub *domain.Subscription, channel *domain.Channel, subscribersCount int) *dto.SubscriptionChannelResponse {
	if sub == nil || channel == nil {
		return nil
	}
	return &dto.SubscriptionChannelResponse{
		ID:               channel.ID,
		UserID:           channel.UserID,
		Name:             channel.Name,
		Description:      channel.Description,
		SubscribersCount: subscribersCount,
		NewVideosCount:   sub.NewVideosCount,
		SubscribedAt:     sub.SubscribedAt,
	}
}

func ToSubscriptionChannelListResponse(subs []*domain.Subscription, channels map[int]*domain.Channel, subscribersCount map[int]int) []*dto.SubscriptionChannelResponse {
	resp := make([]*dto.SubscriptionChannelResponse, 0, len(subs))
	for _, sub := range subs {
		ch := channels[sub.ChannelID]
		if ch == nil {
			continue
		}
		resp = append(resp, ToSubscriptionChannelResponse(sub, ch, subscribersCount[sub.ChannelID]))
	}
	return resp
}

func ToSubscriptionListResponse(subs []*domain.Subscription) []*dto.SubscriptionResponse {
	resp := make([]*dto.SubscriptionResponse, len(subs))
	for i, sub := range subs {
		resp[i] = &dto.SubscriptionResponse{
			ChannelID:      sub.ChannelID,
			ChannelName:    "",
			NewVideosCount: sub.NewVideosCount,
			SubscribedAt:   sub.SubscribedAt,
		}
	}
	return resp
}
