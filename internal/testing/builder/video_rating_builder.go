package builder

import (
	"ZVideo/internal/domain"
	"time"
)

type VideoRatingBuilder struct {
	rating *domain.VideoRating
}

func NewVideoRatingBuilder() *VideoRatingBuilder {
	return &VideoRatingBuilder{
		rating: &domain.VideoRating{
			UserID:  1,
			VideoID: 1,
			Liked:   true,
			RatedAt: time.Now(),
		},
	}
}

func (b *VideoRatingBuilder) WithUserID(id int) *VideoRatingBuilder {
	b.rating.UserID = id
	return b
}

func (b *VideoRatingBuilder) WithVideoID(id int) *VideoRatingBuilder {
	b.rating.VideoID = id
	return b
}

func (b *VideoRatingBuilder) Disliked() *VideoRatingBuilder {
	b.rating.Liked = false
	return b
}

func (b *VideoRatingBuilder) Build() *domain.VideoRating {
	return b.rating
}
