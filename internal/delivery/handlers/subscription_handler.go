package handlers

import (
	"ZVideo/internal/delivery/handlers/mappers"
	"ZVideo/internal/delivery/middleware"
	"ZVideo/internal/delivery/response"
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type SubscriptionHandler struct {
	subSvc service.SubscriptionService
	chlSvc service.ChannelService
}

func NewSubscriptionHandler(
	subSvc service.SubscriptionService,
	chlSvc service.ChannelService,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subSvc: subSvc,
		chlSvc: chlSvc,
	}
}

func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "Subscribe"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to Subscribe",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	channelID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse channel ID in Subscribe",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid channel id")
		return
	}

	logger = logger.With(slog.Int("channel_id", channelID))

	err = h.subSvc.Subscribe(r.Context(), userCtx.UserID, channelID)
	if err != nil {
		logger.WarnContext(r.Context(), "Subscription failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Subscribed successfully"})
}

func (h *SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "Unsubscribe"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to Unsubscribe",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	channelID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to parse channel ID in Unsubscribe",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid channel id")
		return
	}

	logger = logger.With(slog.Int("channel_id", channelID))

	err = h.subSvc.Unsubscribe(r.Context(), userCtx.UserID, channelID)
	if err != nil {
		logger.WarnContext(r.Context(), "Unsubscription failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Unsubscribed successfully"})
}

func (h *SubscriptionHandler) GetUserSubscriptions(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "GetUserSubscriptions"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to GetUserSubscriptions",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	logger = logger.With(slog.Int("user_id", userCtx.UserID))

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit, offset := parsePagination(limitStr, offsetStr)

	logger = logger.With(
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	subs, err := h.subSvc.GetUserSubscriptions(r.Context(), userCtx.UserID, limit, offset)
	if err != nil {
		logger.WarnContext(r.Context(), "Failed to get user subscriptions",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	channelIDs := make([]int, len(subs))
	for i, sub := range subs {
		channelIDs[i] = sub.ChannelID
	}
	channelsMap := make(map[int]*domain.Channel)
	subscribersMap := make(map[int]int)
	for _, id := range channelIDs {
		ch, err := h.chlSvc.GetChannel(r.Context(), id)
		if err != nil {
			logger.WarnContext(r.Context(), "Failed to get channel info for subscription",
				slog.Int("channel_id", id),
				slog.String("error", err.Error()))
		}
		if ch != nil {
			channelsMap[id] = ch
		}

		count, err := h.subSvc.GetSubscribersCount(r.Context(), id)
		if err != nil {
			logger.WarnContext(r.Context(), "Failed to get subscriber count",
				slog.Int("channel_id", id),
				slog.String("error", err.Error()))
		}
		subscribersMap[id] = count
	}

	resp := mappers.ToSubscriptionChannelListResponse(subs, channelsMap, subscribersMap)
	response.RespondWithJSON(w, http.StatusOK, resp)
}
