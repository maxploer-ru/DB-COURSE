package mappers

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainSubscription(model *models.Subscription) *domain.Subscription {
	if model == nil {
		return nil
	}
	return &domain.Subscription{
		UserID:         model.UserID,
		ChannelID:      model.ChannelID,
		NewVideosCount: model.NewVideosCount,
		SubscribedAt:   model.SubscribedAt,
	}
}

func FromDomainSubscription(domainSub *domain.Subscription) *models.Subscription {
	if domainSub == nil {
		return nil
	}
	return &models.Subscription{
		UserID:         domainSub.UserID,
		ChannelID:      domainSub.ChannelID,
		NewVideosCount: domainSub.NewVideosCount,
		SubscribedAt:   domainSub.SubscribedAt,
	}
}

func ToDomainSubscriptionList(models []models.Subscription) []*domain.Subscription {
	subs := make([]*domain.Subscription, len(models))
	for i, m := range models {
		subs[i] = ToDomainSubscription(&m)
	}
	return subs
}
