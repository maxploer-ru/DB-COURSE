package handlers

import (
	"ZVideo/internal/delivery/openapi"
	"ZVideo/internal/delivery/response"
	"ZVideo/internal/domain"
	"net/http"
)

func (h *Handler) ListChannelPlaylists(w http.ResponseWriter, r *http.Request, channelID openapi.ChannelId, params openapi.ListChannelPlaylistsParams) {
	if !validID(int(channelID)) {
		badID(w)
		return
	}
	limit, offset, err := pageParams(params.Limit, params.Offset)
	if err != nil {
		badRequest(w, "INVALID_PAGINATION", err.Error())
		return
	}
	page, err := h.playlist.ListByChannel(r.Context(), int(channelID), limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, playlistPage(page))
}

func (h *Handler) CreatePlaylist(w http.ResponseWriter, r *http.Request, channelID openapi.ChannelId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(channelID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.CreatePlaylistJSONRequestBody](w, r)
	if !ok {
		return
	}
	description := ""
	if body.Description != nil {
		description = *body.Description
	}
	playlist, err := h.playlist.Create(r.Context(), int(channelID), userID, body.Name, description)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusCreated, toPlaylist(playlist))
}

func (h *Handler) ListMyPlaylists(w http.ResponseWriter, r *http.Request, params openapi.ListMyPlaylistsParams) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	limit, offset, err := pageParams(params.Limit, params.Offset)
	if err != nil {
		badRequest(w, "INVALID_PAGINATION", err.Error())
		return
	}
	page, err := h.playlist.GetMyPlaylists(r.Context(), userID, limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, playlistPage(page))
}

func (h *Handler) GetPlaylist(w http.ResponseWriter, r *http.Request, playlistID openapi.PlaylistId) {
	if !validID(int(playlistID)) {
		badID(w)
		return
	}
	playlist, err := h.playlist.GetByID(r.Context(), int(playlistID))
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toPlaylist(playlist))
}

func (h *Handler) UpdatePlaylist(w http.ResponseWriter, r *http.Request, playlistID openapi.PlaylistId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(playlistID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.UpdatePlaylistJSONRequestBody](w, r)
	if !ok {
		return
	}
	playlist, err := h.playlist.Update(r.Context(), int(playlistID), userID, body.Name, body.Description)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toPlaylist(playlist))
}

func (h *Handler) DeletePlaylist(w http.ResponseWriter, r *http.Request, playlistID openapi.PlaylistId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(playlistID)) {
		badID(w)
		return
	}
	if err := h.playlist.Delete(r.Context(), int(playlistID), userID); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) ListPlaylistVideos(w http.ResponseWriter, r *http.Request, playlistID openapi.PlaylistId, params openapi.ListPlaylistVideosParams) {
	if !validID(int(playlistID)) {
		badID(w)
		return
	}
	limit, offset, err := pageParams(params.Limit, params.Offset)
	if err != nil {
		badRequest(w, "INVALID_PAGINATION", err.Error())
		return
	}
	page, err := h.playlist.GetPlaylistItems(r.Context(), int(playlistID), limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	items := make([]openapi.PlaylistItem, 0, len(page.Items))
	for _, item := range page.Items {
		if item != nil {
			items = append(items, toPlaylistItem(item))
		}
	}
	response.RespondWithJSON(w, http.StatusOK, openapi.PlaylistItemPage{
		Items: items, Limit: page.Limit, Offset: page.Offset, TotalCount: page.TotalCount,
	})
}

func (h *Handler) AddVideoToPlaylist(w http.ResponseWriter, r *http.Request, playlistID openapi.PlaylistId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(playlistID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.AddVideoToPlaylistJSONRequestBody](w, r)
	if !ok {
		return
	}
	if !validID(body.VideoId) {
		badID(w)
		return
	}
	if err := h.playlist.AddVideo(r.Context(), int(playlistID), body.VideoId, userID); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) RemoveVideoFromPlaylist(w http.ResponseWriter, r *http.Request, playlistID openapi.PlaylistId, videoID openapi.VideoId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(playlistID)) || !validID(int(videoID)) {
		badID(w)
		return
	}
	if err := h.playlist.RemoveVideo(r.Context(), int(playlistID), int(videoID), userID); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) UpdatePlaylistVideoPosition(w http.ResponseWriter, r *http.Request, playlistID openapi.PlaylistId, videoID openapi.VideoId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(playlistID)) || !validID(int(videoID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.UpdatePlaylistVideoPositionJSONRequestBody](w, r)
	if !ok {
		return
	}
	if body.Position < 1 {
		badRequest(w, "INVALID_POSITION", "Position must be a positive integer")
		return
	}
	if err := h.playlist.UpdateVideoPosition(r.Context(), int(playlistID), int(videoID), userID, body.Position); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func playlistPage(page *domain.PageResponse[*domain.Playlist]) openapi.PlaylistPage {
	items := make([]openapi.Playlist, 0, len(page.Items))
	for _, playlist := range page.Items {
		if playlist != nil {
			items = append(items, toPlaylist(playlist))
		}
	}
	return openapi.PlaylistPage{Items: items, Limit: page.Limit, Offset: page.Offset, TotalCount: page.TotalCount}
}
