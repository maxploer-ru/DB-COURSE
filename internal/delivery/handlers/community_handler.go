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
