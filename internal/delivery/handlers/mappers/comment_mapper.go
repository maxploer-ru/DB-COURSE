package mappers

import (
	"ZVideo/internal/delivery/handlers/dto"
	"ZVideo/internal/domain"
)

func ToCommentResponse(comment *domain.Comment, likes, dislikes int) dto.CommentResponse {
	return dto.CommentResponse{
		ID:        comment.ID,
		UserID:    comment.UserID,
		VideoID:   comment.VideoID,
		Content:   comment.Content,
		Likes:     likes,
		Dislikes:  dislikes,
		CreatedAt: comment.CreatedAt,
	}
}

func ToCommentListResponse(comments []*domain.Comment, stats map[int]struct{ Likes, Dislikes int }, total int64) dto.CommentListResponse {
	resp := make([]dto.CommentResponse, len(comments))
	for i, c := range comments {
		s := stats[c.ID]
		resp[i] = ToCommentResponse(c, s.Likes, s.Dislikes)
	}
	return dto.CommentListResponse{
		Comments: resp,
		Total:    total,
	}
}
