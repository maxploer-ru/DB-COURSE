package handlers

import (
	"ZVideo/internal/delivery/handlers/dto"
	"ZVideo/internal/delivery/handlers/mappers"
	"ZVideo/internal/delivery/middleware"
	"ZVideo/internal/delivery/response"
	"ZVideo/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type PlaylistHandler struct {
	playlistSvc service.PlaylistService
}

func NewPlaylistHandler(playlistSvc service.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{playlistSvc: playlistSvc}
}

// Create creates a playlist.
// @Summary Create playlist
// @Tags Playlists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param channelID path int true "Channel ID"
// @Param request body dto.CreatePlaylistRequest true "Create playlist request"
// @Success 201 {object} dto.PlaylistResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /channels/{channelID}/playlists [post]
func (h *PlaylistHandler) Create(w http.ResponseWriter, r *http.Request) {
	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	channelID, err := strconv.Atoi(chi.URLParam(r, "channelID"))
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid channel id")
		return
	}

	var req dto.CreatePlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	playlist, err := h.playlistSvc.Create(r.Context(), channelID, userCtx.UserID, req.Name, req.Description)
	if err != nil {
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusCreated, mappers.ToPlaylistResponse(playlist))
}

// ListByChannel lists playlists for a channel.
// @Summary List playlists by channel
// @Tags Playlists
// @Produce json
// @Security BearerAuth
// @Param channelID path int true "Channel ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} dto.PlaylistResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /channels/{channelID}/playlists [get]
func (h *PlaylistHandler) ListByChannel(w http.ResponseWriter, r *http.Request) {
	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}
	_ = userCtx

	channelID, err := strconv.Atoi(chi.URLParam(r, "channelID"))
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid channel id")
		return
	}

	limit, offset := parsePagination(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
	playlists, err := h.playlistSvc.ListByChannel(r.Context(), channelID, limit, offset)
	if err != nil {
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, mappers.ToPlaylistListResponse(playlists))
}

// Update updates a playlist.
// @Summary Update playlist
// @Tags Playlists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Playlist ID"
// @Param request body dto.UpdatePlaylistRequest true "Update playlist request"
// @Success 200 {object} dto.PlaylistResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /playlists/{id} [patch]
func (h *PlaylistHandler) Update(w http.ResponseWriter, r *http.Request) {
	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	playlistID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid playlist id")
		return
	}

	var req dto.UpdatePlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	playlist, err := h.playlistSvc.Update(r.Context(), playlistID, userCtx.UserID, req.Name, req.Description)
	if err != nil {
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, mappers.ToPlaylistResponse(playlist))
}

// Delete deletes a playlist.
// @Summary Delete playlist
// @Tags Playlists
// @Produce json
// @Security BearerAuth
// @Param id path int true "Playlist ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /playlists/{id} [delete]
func (h *PlaylistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	playlistID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid playlist id")
		return
	}

	if err := h.playlistSvc.Delete(r.Context(), playlistID, userCtx.UserID); err != nil {
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Playlist deleted successfully"})
}

// AddVideo adds a video to a playlist.
// @Summary Add video to playlist
// @Tags Playlists
// @Produce json
// @Security BearerAuth
// @Param id path int true "Playlist ID"
// @Param videoID path int true "Video ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /playlists/{id}/videos/{videoID} [post]
func (h *PlaylistHandler) AddVideo(w http.ResponseWriter, r *http.Request) {
	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	playlistID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid playlist id")
		return
	}
	videoID, err := strconv.Atoi(chi.URLParam(r, "videoID"))
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid video id")
		return
	}

	if err := h.playlistSvc.AddVideo(r.Context(), playlistID, videoID, userCtx.UserID); err != nil {
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Video added to playlist"})
}

// RemoveVideo removes a video from a playlist.
// @Summary Remove video from playlist
// @Tags Playlists
// @Produce json
// @Security BearerAuth
// @Param id path int true "Playlist ID"
// @Param videoID path int true "Video ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /playlists/{id}/videos/{videoID} [delete]
func (h *PlaylistHandler) RemoveVideo(w http.ResponseWriter, r *http.Request) {
	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	playlistID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid playlist id")
		return
	}
	videoID, err := strconv.Atoi(chi.URLParam(r, "videoID"))
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid video id")
		return
	}

	if err := h.playlistSvc.RemoveVideo(r.Context(), playlistID, videoID, userCtx.UserID); err != nil {
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Video removed from playlist"})
}

// GetByID returns a playlist by ID.
// @Summary Get playlist by ID
// @Tags Playlists
// @Produce json
// @Security BearerAuth
// @Param id path int true "Playlist ID"
// @Success 200 {object} dto.PlaylistResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /playlists/{id} [get]
func (h *PlaylistHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}
	_ = userCtx

	playlistID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid playlist id")
		return
	}

	playlist, err := h.playlistSvc.GetByID(r.Context(), playlistID)
	if err != nil {
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, mappers.ToPlaylistResponse(playlist))
}
