package usecase

import (
	"ZVideo/internal/domain/channel/entity"
	"ZVideo/internal/domain/channel/service"
	"context"
	"fmt"
)

type CreateChannelCommand struct {
	UserID      int
	Name        string
	Description string
}

type CreateChannelResult struct {
	Channel *entity.Channel
}

type CreateChannelUseCase struct {
	channelSvc service.ChannelService
}

func NewCreateChannelUseCase(channelSvc service.ChannelService) *CreateChannelUseCase {
	return &CreateChannelUseCase{channelSvc: channelSvc}
}

func (uc *CreateChannelUseCase) Execute(ctx context.Context, cmd CreateChannelCommand) (*CreateChannelResult, error) {
	ch, err := uc.channelSvc.CreateChannel(ctx, cmd.UserID, cmd.Name, cmd.Description)
	if err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}
	return &CreateChannelResult{Channel: ch}, nil
}
