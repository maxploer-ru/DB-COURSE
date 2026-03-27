package usecase

import (
	"ZVideo/internal/domain/channel/entity"
	"ZVideo/internal/domain/channel/service"
	"context"
	"fmt"
)

type GetChannelQuery struct {
	ChannelID int
}

type GetChannelResult struct {
	Channel *entity.Channel
}

type GetChannelUseCase struct {
	channelSvc service.ChannelService
}

func NewGetChannelUseCase(channelSvc service.ChannelService) *GetChannelUseCase {
	return &GetChannelUseCase{channelSvc: channelSvc}
}

func (uc *GetChannelUseCase) Execute(ctx context.Context, q GetChannelQuery) (*GetChannelResult, error) {
	ch, err := uc.channelSvc.GetChannel(ctx, q.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}
	return &GetChannelResult{Channel: ch}, nil
}
