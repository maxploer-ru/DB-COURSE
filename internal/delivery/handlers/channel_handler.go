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

type ChannelHandler struct {
	chlSvc service.ChannelService
	subSvc service.SubscriptionService
}

func NewChannelHandler(channelService service.ChannelService, subscriptionService service.SubscriptionService) *ChannelHandler {
	return &ChannelHandler{
		chlSvc: channelService,
		subSvc: subscriptionService,
	}
}

func (ch *ChannelHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "CreateChannel"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to CreateChannel",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	req := dto.CreateChannelRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Failed to decode CreateChannel request",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	logger = logger.With(
		slog.Int("user_id", userCtx.UserID),
		slog.String("channel_name", req.ChannelName),
		slog.String("description", req.Description),
	)

	if _, err := ch.chlSvc.CreateChannel(r.Context(), userCtx.UserID, req.ChannelName, req.Description); err != nil {
		logger.WarnContext(r.Context(), "Channel creation failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "Channel created successfully"})
}

func (ch *ChannelHandler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "UpdateChannel"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to UpdateChannel",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	channelID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse channel ID in UpdateChannel",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	req := dto.UpdateChannelRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Failed to decode UpdateChannel request",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	logger = logger.With(
		slog.Int("channel_id", channelID),
	)

	if _, err := ch.chlSvc.UpdateChannel(r.Context(), channelID, userCtx.UserID, req.ChannelName, req.Description); err != nil {
		logger.WarnContext(r.Context(), "Channel update failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Channel updated successfully"})
}

func (ch *ChannelHandler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "DeleteChannel"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to DeleteChannel",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	channelID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse channel ID in DeleteChannel",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	logger = logger.With(
		slog.Int("channel_id", channelID),
		slog.Int("user_id", userCtx.UserID),
	)

	if err := ch.chlSvc.DeleteChannel(r.Context(), channelID, userCtx.UserID); err != nil {
		logger.WarnContext(r.Context(), "Channel delete failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Channel deleted successfully"})
}

func (ch *ChannelHandler) GetChannel(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "GetChannel"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to GetChannel",
			slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	channelID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse channel ID in GetChannel",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	logger = logger.With(
		slog.Int("channel_id", channelID),
	)

	channel, err := ch.chlSvc.GetChannel(r.Context(), channelID)
	if err != nil {
		logger.WarnContext(r.Context(), "Channel get failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	subsCount, err := ch.subSvc.GetSubscribersCount(r.Context(), channelID)
	if err != nil {
		logger.WarnContext(r.Context(), "Subscriber count get failed",
			slog.String("error", err.Error()))
	}

	resp := mappers.ToGetChannelResponse(channel, subsCount)
	response.RespondWithJSON(w, http.StatusOK, resp)

	if err := ch.subSvc.ResetNewVideosCount(r.Context(), userCtx.UserID, channelID); err != nil {
		logger.DebugContext(r.Context(), "No unread counter reset for channel", slog.String("error", err.Error()))
	}
}

func (ch *ChannelHandler) GetMyChannel(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "GetMyChannel"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to GetMyChannel",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	logger = logger.With(
		slog.Int("user_id", userCtx.UserID),
	)

	channel, err := ch.chlSvc.GetChannelByUserID(r.Context(), userCtx.UserID)
	if err != nil {
		logger.WarnContext(r.Context(), "Channel get failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	subsCount, err := ch.subSvc.GetSubscribersCount(r.Context(), channel.ID)
	if err != nil {
		logger.WarnContext(r.Context(), "Subscriber count get failed",
			slog.String("error", err.Error()))
	}

	resp := mappers.ToGetChannelResponse(channel, subsCount)
	response.RespondWithJSON(w, http.StatusOK, resp)
}
