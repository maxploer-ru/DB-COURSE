package mappers

import (
	"ZVideo/internal/delivery/handlers/dto"
	"ZVideo/internal/domain"
)

func ToPlaylistResponse(playlist *domain.Playlist) *dto.PlaylistResponse {
	if playlist == nil {
		return nil
	}

	items := make([]dto.PlaylistItemResponse, 0, len(playlist.Items))
	for _, item := range playlist.Items {
		items = append(items, dto.PlaylistItemResponse{
			VideoID: item.VideoID,
			Number:  item.Number,
		})
	}

	return &dto.PlaylistResponse{
		ID:          playlist.ID,
		ChannelID:   playlist.ChannelID,
		Name:        playlist.Name,
		Description: playlist.Description,
		CreatedAt:   playlist.CreatedAt,
		Items:       items,
	}
}

func ToPlaylistListResponse(playlists []*domain.Playlist) []*dto.PlaylistResponse {
	resp := make([]*dto.PlaylistResponse, 0, len(playlists))
	for _, playlist := range playlists {
		resp = append(resp, ToPlaylistResponse(playlist))
	}
	return resp
}
