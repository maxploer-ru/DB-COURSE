package usecase

import (
	"ZVideo/internal/domain/channel/service"
	"context"
	"fmt"
)

type DeleteChannelCommand struct {
	ChannelID int
	UserID    int
}

type DeleteChannelUseCase struct {
	channelSvc service.ChannelService
}

func NewDeleteChannelUseCase(channelSvc service.ChannelService) *DeleteChannelUseCase {
	return &DeleteChannelUseCase{channelSvc: channelSvc}
}

func (uc *DeleteChannelUseCase) Execute(ctx context.Context, cmd DeleteChannelCommand) error {
	if err := uc.channelSvc.DeleteChannel(ctx, cmd.ChannelID, cmd.UserID); err != nil {
		return fmt.Errorf("delete channel: %w", err)
	}
	return nil
}
