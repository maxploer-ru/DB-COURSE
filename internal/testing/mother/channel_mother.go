package mother

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/testing/builder"
)

type ChannelMother struct{}

func (m *ChannelMother) ValidChannel() *domain.Channel {
	return builder.NewChannelBuilder().Build()
}

func (m *ChannelMother) ChannelForUser(userID int) *domain.Channel {
	return builder.NewChannelBuilder().WithUserID(userID).Build()
}
