package builder

import (
	"ZVideo/internal/domain"
	"time"
)

type CommentRatingBuilder struct {
	rating *domain.CommentRating
}

func NewCommentRatingBuilder() *CommentRatingBuilder {
	return &CommentRatingBuilder{
		rating: &domain.CommentRating{
			UserID:    1,
			CommentID: 1,
			Liked:     true,
			RatedAt:   time.Now(),
		},
	}
}

func (b *CommentRatingBuilder) WithUserID(userID int) *CommentRatingBuilder {
	b.rating.UserID = userID
	return b
}

func (b *CommentRatingBuilder) WithCommentID(commentID int) *CommentRatingBuilder {
	b.rating.CommentID = commentID
	return b
}

func (b *CommentRatingBuilder) Disliked() *CommentRatingBuilder {
	b.rating.Liked = false
	return b
}

func (b *CommentRatingBuilder) Build() *domain.CommentRating {
	return b.rating
}
