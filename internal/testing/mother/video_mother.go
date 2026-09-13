package mother

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/testing/builder"
)

type VideoMother struct{}

func (m *VideoMother) ReadyVideoForChannel(channelID int) *domain.Video {
	return builder.NewVideoBuilder().WithChannelID(channelID).Build()
}

func (m *VideoMother) PendingVideoForChannel(channelID int) *domain.Video {
	return builder.NewVideoBuilder().WithChannelID(channelID).Pending().Build()
}
