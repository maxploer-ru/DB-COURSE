package mappers

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainPlaylist(model *models.Playlist) *domain.Playlist {
	if model == nil {
		return nil
	}

	return &domain.Playlist{
		ID:          model.ID,
		ChannelID:   model.ChannelID,
		Name:        model.Name,
		Description: model.Description,
		CreatedAt:   model.CreatedAt,
	}
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
