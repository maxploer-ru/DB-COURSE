package usecase

import (
	"ZVideo/internal/domain/channel/service"
	"context"
	"fmt"
)

type GetSubscriberCountQuery struct {
	ChannelID int
}

type GetSubscriberCountResult struct {
	Count int
}

type GetSubscriberCountUseCase struct {
	subSvc service.SubscriptionService
}

func NewGetSubscriberCountUseCase(subSvc service.SubscriptionService) *GetSubscriberCountUseCase {
	return &GetSubscriberCountUseCase{subSvc: subSvc}
}

func (uc *GetSubscriberCountUseCase) Execute(ctx context.Context, q GetSubscriberCountQuery) (*GetSubscriberCountResult, error) {
	cnt, err := uc.subSvc.GetSubscriberCount(ctx, q.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("get subscriber count: %w", err)
	}
	return &GetSubscriberCountResult{Count: cnt}, nil
}
