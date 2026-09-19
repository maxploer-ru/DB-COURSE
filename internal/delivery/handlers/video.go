package handlers

import (
	"ZVideo/internal/delivery/openapi"
	"ZVideo/internal/delivery/response"
	"ZVideo/internal/domain"
	"net/http"
	"time"
)

func (h *Handler) ListVideos(w http.ResponseWriter, r *http.Request, params openapi.ListVideosParams) {
	limit, offset, err := pageParams(params.Limit, params.Offset)
	if err != nil {
		badRequest(w, "INVALID_PAGINATION", err.Error())
		return
	}
	page, err := h.video.ListAllVideos(r.Context(), limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, videoPage(page))
}

func (h *Handler) ListMyVideos(w http.ResponseWriter, r *http.Request, params openapi.ListMyVideosParams) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	limit, offset, err := pageParams(params.Limit, params.Offset)
	if err != nil {
		badRequest(w, "INVALID_PAGINATION", err.Error())
		return
	}
	page, err := h.video.ListMyVideos(r.Context(), userID, limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, videoPage(page))
}

func (h *Handler) ListChannelVideos(w http.ResponseWriter, r *http.Request, channelID openapi.ChannelId, params openapi.ListChannelVideosParams) {
	if !validID(int(channelID)) {
		badID(w)
		return
	}
	limit, offset, err := pageParams(params.Limit, params.Offset)
	if err != nil {
		badRequest(w, "INVALID_PAGINATION", err.Error())
		return
	}
	page, err := h.video.ListChannelVideos(r.Context(), int(channelID), limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, videoPage(page))
}

func (h *Handler) InitVideoUpload(w http.ResponseWriter, r *http.Request, channelID openapi.ChannelId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(channelID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.InitVideoUploadJSONRequestBody](w, r)
	if !ok {
		return
	}
	description := ""
	if body.Description != nil {
		description = *body.Description
	}
	video, uploadURL, err := h.video.InitUpload(r.Context(), int(channelID), userID, body.Title, description, body.Filename)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusCreated, openapi.VideoUploadInitResponse{
		Video:           toVideo(video),
		UploadUrl:       uploadURL,
		UploadExpiresAt: time.Now().Add(uploadURLTTL),
	})
}

func (h *Handler) GetVideo(w http.ResponseWriter, r *http.Request, videoID openapi.VideoId) {
	if !validID(int(videoID)) {
		badID(w)
		return
	}
	video, err := h.video.GetVideo(r.Context(), int(videoID))
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toVideo(video))
}

func (h *Handler) UpdateVideo(w http.ResponseWriter, r *http.Request, videoID openapi.VideoId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(videoID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.UpdateVideoJSONRequestBody](w, r)
	if !ok {
		return
	}
	video, err := h.video.UpdateVideo(r.Context(), int(videoID), userID, body.Title, body.Description)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toVideo(video))
}

func (h *Handler) DeleteVideo(w http.ResponseWriter, r *http.Request, videoID openapi.VideoId) {
	userID, role, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(videoID)) {
		badID(w)
		return
	}
	if err := h.video.DeleteVideo(r.Context(), int(videoID), userID, role); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) ConfirmVideoUpload(w http.ResponseWriter, r *http.Request, videoID openapi.VideoId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(videoID)) {
		badID(w)
		return
	}
	if err := h.video.ConfirmUpload(r.Context(), int(videoID), userID); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) GetVideoStreamUrl(w http.ResponseWriter, r *http.Request, videoID openapi.VideoId) {
	if !validID(int(videoID)) {
		badID(w)
		return
	}
	url, err := h.video.GetStreamingPresignedURL(r.Context(), int(videoID))
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, openapi.PresignedUrlResponse{
		Url: url, ExpiresAt: time.Now().Add(streamURLTTL),
	})
}

func (h *Handler) RecordVideoView(w http.ResponseWriter, r *http.Request, videoID openapi.VideoId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(videoID)) {
		badID(w)
		return
	}
	if err := h.videoInteraction.RecordView(r.Context(), userID, int(videoID)); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) RateVideo(w http.ResponseWriter, r *http.Request, videoID openapi.VideoId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(videoID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.RateVideoJSONRequestBody](w, r)
	if !ok {
		return
	}
	if err := h.videoInteraction.Rate(r.Context(), userID, int(videoID), domain.RatingAction(body.Action)); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) GetVideoStats(w http.ResponseWriter, r *http.Request, videoID openapi.VideoId) {
	if !validID(int(videoID)) {
		badID(w)
		return
	}
	if _, err := h.video.GetVideo(r.Context(), int(videoID)); err != nil {
		handleError(w, err)
		return
	}
	stats, err := h.videoInteraction.GetStats(r.Context(), int(videoID))
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, openapi.VideoStats{
		Views: stats.Views, Likes: stats.Likes, Dislikes: stats.Dislikes, Comments: stats.Comments,
	})
}

func videoPage(page *domain.PageResponse[*domain.Video]) openapi.VideoPage {
	items := make([]openapi.Video, 0, len(page.Items))
	for _, video := range page.Items {
		if video != nil {
			items = append(items, toVideo(video))
		}
	}
	return openapi.VideoPage{Items: items, Limit: page.Limit, Offset: page.Offset, TotalCount: page.TotalCount}
}
