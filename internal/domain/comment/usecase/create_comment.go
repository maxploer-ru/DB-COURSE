package usecase

import (
	"ZVideo/internal/domain/comment/entity"
	"ZVideo/internal/domain/comment/service"
	"context"
)

type CreateCommentUseCase struct {
	CommentService service.CommentService
}

func NewCreateCommentUseCase(cs service.CommentService) *CreateCommentUseCase {
	return &CreateCommentUseCase{
		CommentService: cs,
	}
}

func (uc *CreateCommentUseCase) Execute(ctx context.Context, videoID, userID int, text string) (*entity.Comment, error) {
	return uc.CommentService.CreateComment(ctx, userID, videoID, text)
}
