package usecase

import (
	"ZVideo/internal/domain/video/entity"
	"ZVideo/internal/domain/video/service"
	"context"
)

type GetVideoUseCase struct {
	VideoService service.VideoService
}

func NewGetVideoUseCase(vs service.VideoService) *GetVideoUseCase {
	return &GetVideoUseCase{
		VideoService: vs,
	}
}

func (uc *GetVideoUseCase) Execute(ctx context.Context, videoID int, userID *int) (*entity.Video, error) {
	if userID != nil {
		_ = uc.VideoService.RecordViewing(ctx, *userID, videoID)
	}
	return uc.VideoService.GetVideo(ctx, videoID)
}
