package usecase

import (
	"ZVideo/internal/domain/video/entity"
	"ZVideo/internal/domain/video/service"
	"context"
)

type GetChannelVideosUseCase struct {
	VideoService service.VideoService
}

func NewGetChannelVideosUseCase(vs service.VideoService) *GetChannelVideosUseCase {
	return &GetChannelVideosUseCase{
		VideoService: vs,
	}
}

func (uc *GetChannelVideosUseCase) Execute(ctx context.Context, channelID, limit, offset int) ([]*entity.Video, error) {
	return uc.VideoService.GetChannelVideos(ctx, channelID, limit, offset)
}
