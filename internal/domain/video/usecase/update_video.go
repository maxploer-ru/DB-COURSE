package usecase

import (
	"ZVideo/internal/domain/video/entity"
	"ZVideo/internal/domain/video/service"
	"context"
)

type UpdateVideoUseCase struct {
	VideoService service.VideoService
}

func NewUpdateVideoUseCase(vs service.VideoService) *UpdateVideoUseCase {
	return &UpdateVideoUseCase{
		VideoService: vs,
	}
}

func (uc *UpdateVideoUseCase) Execute(ctx context.Context, videoID, userID int, title, description *string) (*entity.Video, error) {
	return uc.VideoService.UpdateVideo(ctx, videoID, userID, title, description)
}
