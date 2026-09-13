package builder

import (
	"ZVideo/internal/domain"
	"time"
)

type ViewingBuilder struct {
	viewing *domain.Viewing
}

func NewViewingBuilder() *ViewingBuilder {
	return &ViewingBuilder{
		viewing: &domain.Viewing{
			UserID:    1,
			VideoID:   1,
			WatchedAt: time.Now(),
		},
	}
}

func (b *ViewingBuilder) WithUserID(id int) *ViewingBuilder {
	b.viewing.UserID = id
	return b
}

func (b *ViewingBuilder) WithVideoID(id int) *ViewingBuilder {
	b.viewing.VideoID = id
	return b
}

func (b *ViewingBuilder) Build() *domain.Viewing {
	return b.viewing
}
