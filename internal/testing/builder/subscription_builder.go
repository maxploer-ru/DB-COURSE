package builder

import (
	"ZVideo/internal/domain"
	"time"
)

type SubscriptionBuilder struct {
	sub *domain.Subscription
}

func NewSubscriptionBuilder() *SubscriptionBuilder {
	return &SubscriptionBuilder{
		sub: &domain.Subscription{
			UserID:         1,
			ChannelID:      2,
			ChannelName:    "Test Channel",
			NewVideosCount: 0,
			SubscribedAt:   time.Now(),
		},
	}
}

func (b *SubscriptionBuilder) WithUserID(id int) *SubscriptionBuilder {
	b.sub.UserID = id
	return b
}

func (b *SubscriptionBuilder) WithChannelID(id int) *SubscriptionBuilder {
	b.sub.ChannelID = id
	return b
}

func (b *SubscriptionBuilder) WithNewVideos(count int) *SubscriptionBuilder {
	b.sub.NewVideosCount = count
	return b
}

func (b *SubscriptionBuilder) Build() *domain.Subscription {
	return b.sub
}
