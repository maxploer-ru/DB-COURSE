package mappers

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainPlaylist(model *models.Playlist) *domain.Playlist {
	if model == nil {
		return nil
	}

	playlist := &domain.Playlist{
		ID:          model.ID,
		ChannelID:   model.ChannelID,
		Name:        model.Name,
		Description: model.Description,
		CreatedAt:   model.CreatedAt,
	}

	if len(model.PlaylistVideos) > 0 {
		playlist.Items = make([]domain.PlaylistItem, 0, len(model.PlaylistVideos))
		for _, item := range model.PlaylistVideos {
			playlist.Items = append(playlist.Items, domain.PlaylistItem{
				VideoID:    item.VideoID,
				VideoTitle: item.Video.Title,
				Number:     item.Number,
			})
		}
	}

	return playlist
}

func FromDomainPlaylist(playlist *domain.Playlist) *models.Playlist {
	if playlist == nil {
		return nil
	}

	return &models.Playlist{
		ID:          playlist.ID,
		ChannelID:   playlist.ChannelID,
		Name:        playlist.Name,
		Description: playlist.Description,
	}
}

func ToDomainPlaylistList(modelsList []models.Playlist) []*domain.Playlist {
	playlists := make([]*domain.Playlist, 0, len(modelsList))
	for i := range modelsList {
		playlists = append(playlists, ToDomainPlaylist(&modelsList[i]))
	}
	return playlists
}
