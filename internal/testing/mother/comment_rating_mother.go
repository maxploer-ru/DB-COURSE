package mother

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/testing/builder"
)

type CommentRatingMother struct{}

func (m *CommentRatingMother) LikeForComment(userID, commentID int) *domain.CommentRating {
	return builder.NewCommentRatingBuilder().
		WithUserID(userID).
		WithCommentID(commentID).
		Build()
}

func (m *CommentRatingMother) DislikeForComment(userID, commentID int) *domain.CommentRating {
	return builder.NewCommentRatingBuilder().
		WithUserID(userID).
		WithCommentID(commentID).
		Disliked().
		Build()
}
