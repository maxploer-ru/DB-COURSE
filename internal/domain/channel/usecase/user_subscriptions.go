package usecase

import (
	"ZVideo/internal/domain/channel/entity"
	"ZVideo/internal/domain/channel/service"
	"context"
	"fmt"
)

type GetUserSubscriptionsQuery struct {
	UserID int
}

type GetUserSubscriptionsResult struct {
	Subscriptions []*entity.Subscription
}

type GetUserSubscriptionsUseCase struct {
	subSvc service.SubscriptionService
}

func NewGetUserSubscriptionsUseCase(subSvc service.SubscriptionService) *GetUserSubscriptionsUseCase {
	return &GetUserSubscriptionsUseCase{subSvc: subSvc}
}

func (uc *GetUserSubscriptionsUseCase) Execute(ctx context.Context, q GetUserSubscriptionsQuery) (*GetUserSubscriptionsResult, error) {
	subs, err := uc.subSvc.GetUserSubscriptions(ctx, q.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user subscriptions: %w", err)
	}
	return &GetUserSubscriptionsResult{Subscriptions: subs}, nil
}
