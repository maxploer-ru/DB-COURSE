package usecase

import (
	"ZVideo/internal/domain/channel/service"
	"context"
	"fmt"
)

type UnsubscribeCommand struct {
	UserID    int
	ChannelID int
}

type UnsubscribeUseCase struct {
	subSvc service.SubscriptionService
}

func NewUnsubscribeUseCase(subSvc service.SubscriptionService) *UnsubscribeUseCase {
	return &UnsubscribeUseCase{subSvc: subSvc}
}

func (uc *UnsubscribeUseCase) Execute(ctx context.Context, cmd UnsubscribeCommand) error {
	if err := uc.subSvc.Unsubscribe(ctx, cmd.UserID, cmd.ChannelID); err != nil {
		return fmt.Errorf("unsubscribe: %w", err)
	}
	return nil
}
