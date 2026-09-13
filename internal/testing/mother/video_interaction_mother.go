package mother

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/testing/builder"
)

type VideoInteractionMother struct{}

func (m *VideoInteractionMother) LikedRating() *domain.VideoRating {
	return builder.NewVideoRatingBuilder().Build()
}

func (m *VideoInteractionMother) DislikedRating() *domain.VideoRating {
	return builder.NewVideoRatingBuilder().Disliked().Build()
}
