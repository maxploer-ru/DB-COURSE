package handlers

import (
	"ZVideo/internal/delivery/middleware"
	"ZVideo/internal/delivery/response"
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type VideoInteractionHandler struct {
	interactionSvc service.VideoInteractionService
}

func NewVideoInteractionHandler(interactionSvc service.VideoInteractionService) *VideoInteractionHandler {
	return &VideoInteractionHandler{interactionSvc: interactionSvc}
}

func (h *VideoInteractionHandler) Like(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "LikeVideo"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to LikeVideo",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}
	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	videoID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse video ID in LikeVideo",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid video id")
		return
	}
	logger = logger.With(slog.Int("video_id", videoID))

	err = h.interactionSvc.Like(r.Context(), userCtx.UserID, videoID)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to like video",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "liked"})
}

func (h *VideoInteractionHandler) Dislike(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "DislikeVideo"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to DislikeVideo",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}
	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	videoID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse video ID in DislikeVideo",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid video id")
		return
	}
	logger = logger.With(slog.Int("video_id", videoID))

	err = h.interactionSvc.Dislike(r.Context(), userCtx.UserID, videoID)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to dislike video",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "disliked"})
}

func (h *VideoInteractionHandler) RemoveRating(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "RemoveVideoRating"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to RemoveVideoRating",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}
	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	videoID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse video ID in RemoveVideoRating",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid video id")
		return
	}
	logger = logger.With(slog.Int("video_id", videoID))

	err = h.interactionSvc.RemoveRating(r.Context(), userCtx.UserID, videoID)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to remove video rating",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "rating removed"})
}
