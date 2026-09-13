package mother

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/testing/builder"
)

type ViewingMother struct{}

func (m *ViewingMother) ValidViewing() *domain.Viewing {
	return builder.NewViewingBuilder().Build()
}
