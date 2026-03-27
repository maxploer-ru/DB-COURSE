package usecase

import (
	"ZVideo/internal/domain/channel/entity"
	"ZVideo/internal/domain/channel/service"
	"context"
	"fmt"
)

type GetChannelSubscribersQuery struct {
	ChannelID int
}

type GetChannelSubscribersResult struct {
	Subscribers []*entity.Subscription
}

type GetChannelSubscribersUseCase struct {
	subSvc service.SubscriptionService
}

func NewGetChannelSubscribersUseCase(subSvc service.SubscriptionService) *GetChannelSubscribersUseCase {
	return &GetChannelSubscribersUseCase{subSvc: subSvc}
}

func (uc *GetChannelSubscribersUseCase) Execute(ctx context.Context, q GetChannelSubscribersQuery) (*GetChannelSubscribersResult, error) {
	subs, err := uc.subSvc.GetChannelSubscribers(ctx, q.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("get channel subscribers: %w", err)
	}
	return &GetChannelSubscribersResult{Subscribers: subs}, nil
}
