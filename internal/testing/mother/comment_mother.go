package mother

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/testing/builder"
)

type CommentMother struct{}

func (m *CommentMother) CommentForVideo(userID, videoID int) *domain.Comment {
	return builder.NewCommentBuilder().
		WithUserID(userID).
		WithVideoID(videoID).
		Build()
}
