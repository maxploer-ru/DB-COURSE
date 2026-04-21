package handlers

import (
	"ZVideo/internal/delivery/handlers/dto"
	"ZVideo/internal/delivery/handlers/mappers"
	"ZVideo/internal/delivery/middleware"
	"ZVideo/internal/delivery/response"
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type VideoHandler struct {
	videoSvc       service.VideoService
	interactionSvc service.VideoInteractionService
}

func NewVideoHandler(videoService service.VideoService, interactionService service.VideoInteractionService) *VideoHandler {
	return &VideoHandler{
		videoSvc:       videoService,
		interactionSvc: interactionService,
	}
}

func (h *VideoHandler) CreateVideo(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "CreateVideo"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to CreateVideo",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	var req dto.CreateVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Failed to decode CreateVideo request",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	logger = logger.With(
		slog.Int("channel_id", req.ChannelID),
		slog.String("title", req.Title),
		slog.String("file_key", req.FileKey),
	)

	video, err := h.videoSvc.CreateVideo(r.Context(), req.ChannelID, userCtx.UserID, req.Title, req.Description, req.FileKey)
	if err != nil {
		logger.WarnContext(r.Context(), "Video creation failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	logger = logger.With(slog.Int("video_id", video.ID))

	stats := &domain.VideoStats{
		Views:    0,
		Likes:    0,
		Dislikes: 0,
	}
	resp := mappers.ToVideoResponse(video, stats)
	response.RespondWithJSON(w, http.StatusCreated, resp)
}

func (h *VideoHandler) GetVideo(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "GetVideo"))

	videoID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse video ID in GetVideo",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid video id")
		return
	}

	logger = logger.With(slog.Int("video_id", videoID))

	video, err := h.videoSvc.GetVideo(r.Context(), videoID)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to get video",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	stats, err := h.interactionSvc.GetStats(r.Context(), videoID)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to get video stats",
			slog.String("error", err.Error()))
	}

	resp := mappers.ToVideoResponse(video, stats)
	response.RespondWithJSON(w, http.StatusOK, resp)
}

func (h *VideoHandler) UpdateVideo(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "UpdateVideo"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to UpdateVideo",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	videoID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse video ID in UpdateVideo",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid video id")
		return
	}

	logger = logger.With(slog.Int("video_id", videoID))

	var req dto.UpdateVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Failed to decode UpdateVideo request",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	video, err := h.videoSvc.UpdateVideo(r.Context(), videoID, userCtx.UserID, req.Title, req.Description)
	if err != nil {
		logger.WarnContext(r.Context(), "Video update failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	stats, err := h.interactionSvc.GetStats(r.Context(), videoID)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to get video stats after update",
			slog.String("error", err.Error()))
	}

	resp := mappers.ToVideoResponse(video, stats)
	response.RespondWithJSON(w, http.StatusOK, resp)
}

func (h *VideoHandler) DeleteVideo(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "DeleteVideo"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to DeleteVideo",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	videoID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse video ID in DeleteVideo",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid video id")
		return
	}

	logger = logger.With(slog.Int("video_id", videoID))

	if err := h.videoSvc.DeleteVideo(r.Context(), videoID, userCtx.UserID); err != nil {
		logger.WarnContext(r.Context(), "Video deletion failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Video deleted successfully"})
}

func (h *VideoHandler) ListMyVideos(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "ListMyVideos"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to ListMyVideos",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit, offset := parsePagination(limitStr, offsetStr)

	logger = logger.With(slog.Int("limit", limit), slog.Int("offset", offset))

	videos, err := h.videoSvc.ListMyVideos(r.Context(), userCtx.UserID, limit, offset)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to list user videos",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	stats := make([]*domain.VideoStats, len(videos))
	for i, video := range videos {
		stat, err := h.interactionSvc.GetStats(r.Context(), video.ID)
		if err != nil {
			logger.WarnContext(r.Context(), "Failed to get stats for video",
				slog.Int("video_id", video.ID),
				slog.String("error", err.Error()))
		}
		stats[i] = stat
	}

	resp := mappers.ToVideoListResponse(videos, stats)
	response.RespondWithJSON(w, http.StatusOK, resp)
}

func (h *VideoHandler) List(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "ListVideos"))

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit, offset := parsePagination(limitStr, offsetStr)

	logger = logger.With(slog.Int("limit", limit), slog.Int("offset", offset))

	videos, err := h.videoSvc.ListAllVideos(r.Context(), limit, offset)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to list all videos",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	stats := make([]*domain.VideoStats, len(videos))
	for i, video := range videos {
		stat, err := h.interactionSvc.GetStats(r.Context(), video.ID)
		if err != nil {
			logger.WarnContext(r.Context(), "Failed to get stats for video",
				slog.Int("video_id", video.ID),
				slog.String("error", err.Error()))
		}
		stats[i] = stat
	}

	resp := mappers.ToVideoListResponse(videos, stats)
	response.RespondWithJSON(w, http.StatusOK, resp)
}

func (h *VideoHandler) ListChannelVideos(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "ListChannelVideos"))

	channelID, err := strconv.Atoi(chi.URLParam(r, "channelID"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse channel ID in ListChannelVideos",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid channel id")
		return
	}
	logger = logger.With(slog.Int("channel_id", channelID))

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit, offset := parsePagination(limitStr, offsetStr)

	logger = logger.With(slog.Int("limit", limit), slog.Int("offset", offset))

	videos, err := h.videoSvc.ListChannelVideos(r.Context(), channelID, limit, offset)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to list channel videos",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	stats := make([]*domain.VideoStats, len(videos))
	for i, video := range videos {
		stat, err := h.interactionSvc.GetStats(r.Context(), video.ID)
		if err != nil {
			logger.WarnContext(r.Context(), "Failed to get stats for video",
				slog.Int("video_id", video.ID),
				slog.String("error", err.Error()))
		}
		stats[i] = stat
	}

	resp := mappers.ToVideoListResponse(videos, stats)
	response.RespondWithJSON(w, http.StatusOK, resp)
}

func (h *VideoHandler) GetUploadPresignedURL(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "GetUploadPresignedURL"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to GetUploadPresignedURL",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}
	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	var req dto.UploadPresignedURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Failed to decode UploadPresignedURL request",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	logger = logger.With(slog.Int("channel_id", req.ChannelID), slog.String("filename", req.Filename))

	url, fileKey, err := h.videoSvc.GetUploadPresignedURL(r.Context(), req.ChannelID, userCtx.UserID, req.Filename)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to generate upload presigned URL",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	resp := dto.UploadPresignedURLResponse{
		URL:     url,
		FileKey: fileKey,
	}
	response.RespondWithJSON(w, http.StatusOK, resp)
}

func (h *VideoHandler) GetStreamingPresignedURL(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "GetStreamingPresignedURL"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to GetStreamingPresignedURL",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}
	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	videoID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse video ID in GetStreamingPresignedURL",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid video id")
		return
	}
	logger = logger.With(slog.Int("video_id", videoID))

	url, err := h.videoSvc.GetStreamingPresignedURL(r.Context(), videoID)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to generate streaming presigned URL",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.interactionSvc.RecordView(ctx, userCtx.UserID, videoID); err != nil {
			domain.GetLogger(ctx).WarnContext(ctx, "Failed to record view in background",
				slog.Int("user_id", userCtx.UserID),
				slog.Int("video_id", videoID),
				slog.String("error", err.Error()))
		}
	}()

	resp := dto.StreamingURLResponse{URL: url}
	response.RespondWithJSON(w, http.StatusOK, resp)
}

func parsePagination(limitStr, offsetStr string) (int, int) {
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	return limit, offset
}
