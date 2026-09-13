package mother

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/testing/builder"
)

type SubscriptionMother struct{}

func (m *SubscriptionMother) ValidSubscription() *domain.Subscription {
	return builder.NewSubscriptionBuilder().Build()
}

func (m *SubscriptionMother) SelfSubscription() *domain.Subscription {
	return builder.NewSubscriptionBuilder().WithUserID(1).WithChannelID(1).Build()
}
