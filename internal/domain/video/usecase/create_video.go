package usecase

import (
	"ZVideo/internal/domain/video/entity"
	"ZVideo/internal/domain/video/service"
	"context"
)

type CreateVideoUseCase struct {
	VideoService service.VideoService
}

func NewCreateVideoUseCase(vs service.VideoService) *CreateVideoUseCase {
	return &CreateVideoUseCase{
		VideoService: vs,
	}
}

func (uc *CreateVideoUseCase) Execute(ctx context.Context, channelID, userID int, title, description, filepath string) (*entity.Video, error) {
	return uc.VideoService.CreateVideo(ctx, channelID, userID, title, description, filepath)
}
