package usecase

import (
	"ZVideo/internal/domain/video/service"
	"context"
)

type DeleteVideoUseCase struct {
	VideoService service.VideoService
}

func NewDeleteVideoUseCase(vs service.VideoService) *DeleteVideoUseCase {
	return &DeleteVideoUseCase{
		VideoService: vs,
	}
}

func (uc *DeleteVideoUseCase) Execute(ctx context.Context, videoID, userID int) error {
	return uc.VideoService.DeleteVideo(ctx, videoID, userID)
}
