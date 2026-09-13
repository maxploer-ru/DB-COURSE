package mother

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/testing/builder"
)

type PlaylistMother struct{}

func (m *PlaylistMother) PlaylistForChannel(channelID int) *domain.Playlist {
	return builder.NewPlaylistBuilder().WithChannelID(channelID).Build()
}
