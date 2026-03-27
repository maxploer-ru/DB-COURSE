package usecase

import (
	"ZVideo/internal/domain/channel/service"
	"context"
	"fmt"
)

type SubscribeCommand struct {
	UserID    int
	ChannelID int
}

type SubscribeUseCase struct {
	subSvc service.SubscriptionService
}

func NewSubscribeUseCase(subSvc service.SubscriptionService) *SubscribeUseCase {
	return &SubscribeUseCase{subSvc: subSvc}
}

func (uc *SubscribeUseCase) Execute(ctx context.Context, cmd SubscribeCommand) error {
	if err := uc.subSvc.Subscribe(ctx, cmd.UserID, cmd.ChannelID); err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}
	return nil
}
