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

type CommunityHandler struct {
	communitySvc service.CommunityService
	subSvc       service.SubscriptionService
}

func NewCommunityHandler(communitySvc service.CommunityService, subscriptionService service.SubscriptionService) *CommunityHandler {
	return &CommunityHandler{
		communitySvc: communitySvc,
		subSvc:       subscriptionService,
	}
}

// GetChannelCommunity returns community for a channel.
// @Summary Get channel community
// @Tags Community
// @Produce json
// @Security BearerAuth
// @Param channelID path int true "Channel ID"
// @Success 200 {object} dto.CommunityResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /channels/{channelID}/community [get]
func (h *CommunityHandler) GetChannelCommunity(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "GetChannelCommunity"))

	channelID, err := strconv.Atoi(chi.URLParam(r, "channelID"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse channel ID in GetChannelCommunity", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid channel id")
		return
	}

	community, err := h.communitySvc.GetChannelCommunity(r.Context(), channelID)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to load channel community", slog.String("error", err.Error()))
		response.HandleDomainError(w, err)
		return
	}

	subsCount, err := h.subSvc.GetSubscribersCount(r.Context(), channelID)
	if err != nil {
		logger.DebugContext(r.Context(), "Failed to load subscriber count", slog.String("error", err.Error()))
	}

	response.RespondWithJSON(w, http.StatusOK, mappers.ToCommunityResponse(community, subsCount))
}

// GetMyCommunity returns community for the current user's channel.
// @Summary Get my community
// @Tags Community
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.CommunityResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /channels/me/community [get]
func (h *CommunityHandler) GetMyCommunity(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "GetMyCommunity"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to GetMyCommunity", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	community, err := h.communitySvc.GetMyCommunity(r.Context(), userCtx.UserID)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to load my community", slog.String("error", err.Error()))
		response.HandleDomainError(w, err)
		return
	}

	subsCount, err := h.subSvc.GetSubscribersCount(r.Context(), community.Channel.ID)
	if err != nil {
		logger.DebugContext(r.Context(), "Failed to load subscriber count", slog.String("error", err.Error()))
	}

	response.RespondWithJSON(w, http.StatusOK, mappers.ToCommunityResponse(community, subsCount))
}

// CreatePost creates a community post.
// @Summary Create community post
// @Tags Community
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param channelID path int true "Channel ID"
// @Param request body dto.CreateCommunityPostRequest true "Create post request"
// @Success 201 {object} dto.CommunityPostResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /channels/{channelID}/community/posts [post]
func (h *CommunityHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "CreateCommunityPost"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to CreateCommunityPost", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	channelID, err := strconv.Atoi(chi.URLParam(r, "channelID"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse channel ID in CreateCommunityPost", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid channel id")
		return
	}

	var req dto.CreateCommunityPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Failed to decode CreateCommunityPost request", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	post, err := h.communitySvc.CreatePost(r.Context(), channelID, userCtx.UserID, req.Content)
	if err != nil {
		logger.WarnContext(r.Context(), "Community post creation failed", slog.String("error", err.Error()))
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusCreated, mappers.ToCommunityPostResponse(post, nil))
}

// UpdatePost updates a community post.
// @Summary Update community post
// @Tags Community
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param postID path int true "Post ID"
// @Param request body dto.UpdateCommunityPostRequest true "Update post request"
// @Success 200 {object} dto.CommunityPostResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /community/posts/{postID} [patch]
func (h *CommunityHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "UpdateCommunityPost"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to UpdateCommunityPost", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	postID, err := strconv.Atoi(chi.URLParam(r, "postID"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse post ID in UpdateCommunityPost", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid post id")
		return
	}

	var req dto.UpdateCommunityPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Failed to decode UpdateCommunityPost request", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	post, err := h.communitySvc.UpdatePost(r.Context(), postID, userCtx.UserID, req.Content)
	if err != nil {
		logger.WarnContext(r.Context(), "Community post update failed", slog.String("error", err.Error()))
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, mappers.ToCommunityPostResponse(post, nil))
}

// DeletePost deletes a community post.
// @Summary Delete community post
// @Tags Community
// @Produce json
// @Security BearerAuth
// @Param postID path int true "Post ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /community/posts/{postID} [delete]
func (h *CommunityHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "DeleteCommunityPost"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to DeleteCommunityPost", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	postID, err := strconv.Atoi(chi.URLParam(r, "postID"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse post ID in DeleteCommunityPost", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid post id")
		return
	}

	if err := h.communitySvc.DeletePost(r.Context(), postID, userCtx.UserID); err != nil {
		logger.WarnContext(r.Context(), "Community post deletion failed", slog.String("error", err.Error()))
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Community post deleted successfully"})
}

// CreateComment creates a comment on a community post.
// @Summary Create community comment
// @Tags Community
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param postID path int true "Post ID"
// @Param request body dto.CreateCommunityCommentRequest true "Create comment request"
// @Success 201 {object} dto.CommunityCommentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /community/posts/{postID}/comments [post]
func (h *CommunityHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "CreateCommunityComment"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to CreateCommunityComment", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	postID, err := strconv.Atoi(chi.URLParam(r, "postID"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse post ID in CreateCommunityComment", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid post id")
		return
	}

	var req dto.CreateCommunityCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Failed to decode CreateCommunityComment request", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	comment, err := h.communitySvc.CreateComment(r.Context(), postID, userCtx.UserID, req.Content)
	if err != nil {
		logger.WarnContext(r.Context(), "Community comment creation failed", slog.String("error", err.Error()))
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusCreated, mappers.ToCommunityCommentResponse(comment))
}

// UpdateComment updates a community comment.
// @Summary Update community comment
// @Tags Community
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Comment ID"
// @Param request body dto.UpdateCommentRequest true "Update comment request"
// @Success 200 {object} dto.CommunityCommentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /community/comments/{id} [patch]
func (h *CommunityHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "UpdateCommunityComment"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to UpdateCommunityComment", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	commentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse comment ID in UpdateCommunityComment", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid comment id")
		return
	}

	var req dto.UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Failed to decode UpdateCommunityComment request", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	comment, err := h.communitySvc.UpdateComment(r.Context(), commentID, userCtx.UserID, req.Content)
	if err != nil {
		logger.WarnContext(r.Context(), "Community comment update failed", slog.String("error", err.Error()))
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, mappers.ToCommunityCommentResponse(comment))
}

// DeleteComment deletes a community comment.
// @Summary Delete community comment
// @Tags Community
// @Produce json
// @Security BearerAuth
// @Param id path int true "Comment ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /community/comments/{id} [delete]
func (h *CommunityHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "DeleteCommunityComment"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to DeleteCommunityComment", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	commentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse comment ID in DeleteCommunityComment", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid comment id")
		return
	}

	if err := h.communitySvc.DeleteComment(r.Context(), commentID, userCtx.UserID); err != nil {
		logger.WarnContext(r.Context(), "Community comment deletion failed", slog.String("error", err.Error()))
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Community comment deleted successfully"})
}
