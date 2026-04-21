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

type CommentInteractionHandler struct {
	interactionSvc service.CommentInteractionService
}

func NewCommentInteractionHandler(interactionSvc service.CommentInteractionService) *CommentInteractionHandler {
	return &CommentInteractionHandler{interactionSvc: interactionSvc}
}

func (h *CommentInteractionHandler) Like(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "LikeComment"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to LikeComment",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	commentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse comment ID in LikeComment",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid comment id")
		return
	}

	logger = logger.With(slog.Int("comment_id", commentID))

	err = h.interactionSvc.Like(r.Context(), userCtx.UserID, commentID)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to like comment",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "liked"})
}

func (h *CommentInteractionHandler) Dislike(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "DislikeComment"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to DislikeComment",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	commentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse comment ID in DislikeComment",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid comment id")
		return
	}

	logger = logger.With(slog.Int("comment_id", commentID))

	err = h.interactionSvc.Dislike(r.Context(), userCtx.UserID, commentID)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to dislike comment",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "disliked"})
}

func (h *CommentInteractionHandler) RemoveRating(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "RemoveCommentRating"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to RemoveCommentRating",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}
	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	commentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse comment ID in RemoveCommentRating",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid comment id")
		return
	}
	logger = logger.With(slog.Int("comment_id", commentID))

	err = h.interactionSvc.RemoveRating(r.Context(), userCtx.UserID, commentID)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to remove comment rating",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "rating removed"})
}
