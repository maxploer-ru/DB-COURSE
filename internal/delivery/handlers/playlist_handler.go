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
