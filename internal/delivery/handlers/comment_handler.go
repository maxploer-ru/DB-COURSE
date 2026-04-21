package handlers

import (
	"ZVideo/internal/delivery/handlers/dto"
	"ZVideo/internal/delivery/handlers/mappers"
	"ZVideo/internal/delivery/middleware"
	"ZVideo/internal/delivery/response"
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CommentHandler struct {
	commentSvc     service.CommentService
	interactionSvc service.CommentInteractionService
}

func NewCommentHandler(
	commentSvc service.CommentService,
	interactionSvc service.CommentInteractionService,
) *CommentHandler {
	return &CommentHandler{
		commentSvc:     commentSvc,
		interactionSvc: interactionSvc,
	}
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "CreateComment"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to CreateComment",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	videoID, err := strconv.Atoi(chi.URLParam(r, "videoID"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse video ID in CreateComment",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid video id")
		return
	}

	logger = logger.With(slog.Int("video_id", videoID))

	var req dto.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Failed to decode CreateComment request",
			slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	comment, err := h.commentSvc.Create(r.Context(), userCtx.UserID, videoID, req.Content)
	if err != nil {
		logger.WarnContext(r.Context(), "Comment creation failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	logger = logger.With(slog.Int("comment_id", comment.ID))

	likes, dislikes, err := h.interactionSvc.GetStats(r.Context(), comment.ID)
	if err != nil {
		logger.WarnContext(r.Context(), "Interaction service failed",
			slog.String("error", err.Error()))
	}

	resp := mappers.ToCommentResponse(comment, int(likes), int(dislikes))
	response.RespondWithJSON(w, http.StatusCreated, resp)
}

func (h *CommentHandler) List(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "ListComments"))

	videoID, err := strconv.Atoi(chi.URLParam(r, "videoID"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse video ID in ListComments",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid video id")
		return
	}

	logger = logger.With(slog.Int("video_id", videoID))

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit, offset := parsePagination(limitStr, offsetStr)

	logger = logger.With(
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	comments, err := h.commentSvc.ListByVideo(r.Context(), videoID, limit, offset)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to list comments",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	statsMap := make(map[int]struct{ Likes, Dislikes int })
	for _, c := range comments {
		likes, dislikes, _ := h.interactionSvc.GetStats(r.Context(), c.ID)
		statsMap[c.ID] = struct{ Likes, Dislikes int }{Likes: int(likes), Dislikes: int(dislikes)}
	}

	total, err := h.commentSvc.GetCount(r.Context(), videoID)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to get comment count",
			slog.String("error", err.Error()))
	}

	resp := mappers.ToCommentListResponse(comments, statsMap, total)
	response.RespondWithJSON(w, http.StatusOK, resp)
}

func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "UpdateComment"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to UpdateComment",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	commentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse comment ID in UpdateComment",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid comment id")
		return
	}

	logger = logger.With(slog.Int("comment_id", commentID))

	var req dto.UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Failed to decode UpdateComment request",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	comment, err := h.commentSvc.Update(r.Context(), commentID, userCtx.UserID, req.Content)
	if err != nil {
		logger.WarnContext(r.Context(), "Comment update failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	likes, dislikes, err := h.interactionSvc.GetStats(r.Context(), comment.ID)
	if err != nil {
		logger.WarnContext(r.Context(), "Interaction service failed",
			slog.String("error", err.Error()))
	}

	resp := mappers.ToCommentResponse(comment, int(likes), int(dislikes))
	response.RespondWithJSON(w, http.StatusOK, resp)
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "DeleteComment"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to DeleteComment",
			slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	commentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse comment ID in DeleteComment",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid comment id")
		return
	}

	logger = logger.With(slog.Int("comment_id", commentID))

	err = h.commentSvc.Delete(r.Context(), commentID, userCtx.UserID, userCtx.Role)
	if err != nil {
		logger.WarnContext(r.Context(), "Comment deletion failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Comment deleted successfully"})
}
