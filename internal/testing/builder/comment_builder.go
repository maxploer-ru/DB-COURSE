package builder

import (
	"ZVideo/internal/domain"
	"time"
)

type CommentBuilder struct {
	comment *domain.Comment
}

func NewCommentBuilder() *CommentBuilder {
	return &CommentBuilder{
		comment: &domain.Comment{
			ID:        1,
			UserID:    1,
			VideoID:   1,
			Content:   "Test comment content",
			CreatedAt: time.Now(),
		},
	}
}

func (b *CommentBuilder) WithID(id int) *CommentBuilder {
	b.comment.ID = id
	return b
}

func (b *CommentBuilder) WithUserID(userID int) *CommentBuilder {
	b.comment.UserID = userID
	return b
}

func (b *CommentBuilder) WithVideoID(videoID int) *CommentBuilder {
	b.comment.VideoID = videoID
	return b
}

func (b *CommentBuilder) Build() *domain.Comment {
	return b.comment
}
