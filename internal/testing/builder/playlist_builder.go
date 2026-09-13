package builder

import (
	"ZVideo/internal/domain"
	"time"
)

type PlaylistBuilder struct {
	playlist *domain.Playlist
}

func NewPlaylistBuilder() *PlaylistBuilder {
	return &PlaylistBuilder{
		playlist: &domain.Playlist{
			ID:          1,
			ChannelID:   1,
			Name:        "test_playlist",
			Description: "desc",
			CreatedAt:   time.Now(),
		},
	}
}

func (b *PlaylistBuilder) WithID(id int) *PlaylistBuilder {
	b.playlist.ID = id
	return b
}

func (b *PlaylistBuilder) WithChannelID(channelID int) *PlaylistBuilder {
	b.playlist.ChannelID = channelID
	return b
}

func (b *PlaylistBuilder) Build() *domain.Playlist {
	return b.playlist
}
