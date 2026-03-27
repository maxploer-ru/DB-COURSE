package usecase

import (
	"ZVideo/internal/domain/channel/entity"
	"ZVideo/internal/domain/channel/service"
	"context"
	"fmt"
)

type UpdateChannelCommand struct {
	ChannelID   int
	UserID      int
	Name        *string
	Description *string
}

type UpdateChannelResult struct {
	Channel *entity.Channel
}

type UpdateChannelUseCase struct {
	channelSvc service.ChannelService
}

func NewUpdateChannelUseCase(channelSvc service.ChannelService) *UpdateChannelUseCase {
	return &UpdateChannelUseCase{channelSvc: channelSvc}
}

func (uc *UpdateChannelUseCase) Execute(ctx context.Context, cmd UpdateChannelCommand) (*UpdateChannelResult, error) {
	ch, err := uc.channelSvc.UpdateChannel(ctx, cmd.ChannelID, cmd.Name, cmd.Description)
	if err != nil {
		return nil, fmt.Errorf("update channel: %w", err)
	}
	if ch.UserID != cmd.UserID {
		return nil, fmt.Errorf("not authorized")
	}
	return &UpdateChannelResult{Channel: ch}, nil
}
