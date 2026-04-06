package usecase

import (
	"ZVideo/internal/domain/comment/entity"
	"ZVideo/internal/domain/comment/service"
	"context"
)

type GetVideoCommentsUseCase struct {
	CommentService service.CommentService
}

func NewGetVideoCommentsUseCase(cs service.CommentService) *GetVideoCommentsUseCase {
	return &GetVideoCommentsUseCase{
		CommentService: cs,
	}
}

func (uc *GetVideoCommentsUseCase) Execute(ctx context.Context, videoID, limit, offset int) ([]*entity.Comment, error) {
	return uc.CommentService.GetCommentsByVideo(ctx, videoID, limit, offset)
}
