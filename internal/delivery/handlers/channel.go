package handlers

import (
	"ZVideo/internal/delivery/openapi"
	"ZVideo/internal/delivery/response"
	"net/http"
)

func (h *Handler) ListChannels(w http.ResponseWriter, r *http.Request, params openapi.ListChannelsParams) {
	limit, offset, err := pageParams(params.Limit, params.Offset)
	if err != nil {
		badRequest(w, "INVALID_PAGINATION", err.Error())
		return
	}
	page, err := h.channel.ListChannels(r.Context(), limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	items := make([]openapi.Channel, 0, len(page.Items))
	for _, channel := range page.Items {
		if channel != nil {
			items = append(items, toChannel(channel))
		}
	}
	response.RespondWithJSON(w, http.StatusOK, openapi.ChannelPage{
		Items: items, Limit: page.Limit, Offset: page.Offset, TotalCount: page.TotalCount,
	})
}

func (h *Handler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	body, ok := decodeJSON[openapi.CreateChannelJSONRequestBody](w, r)
	if !ok {
		return
	}
	description := ""
	if body.Description != nil {
		description = *body.Description
	}
	channel, err := h.channel.CreateChannel(r.Context(), userID, body.Name, description)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusCreated, toChannel(channel))
}

func (h *Handler) GetMyChannel(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	channel, err := h.channel.GetChannelByUserID(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toChannel(channel))
}

func (h *Handler) GetChannel(w http.ResponseWriter, r *http.Request, channelID openapi.ChannelId) {
	if !validID(int(channelID)) {
		badID(w)
		return
	}
	channel, err := h.channel.GetChannel(r.Context(), int(channelID))
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toChannel(channel))
}

func (h *Handler) GetChannelByName(w http.ResponseWriter, r *http.Request, name string) {
	channel, err := h.channel.GetChannelByName(r.Context(), name)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toChannel(channel))
}

func (h *Handler) UpdateChannel(w http.ResponseWriter, r *http.Request, channelID openapi.ChannelId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(channelID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.UpdateChannelJSONRequestBody](w, r)
	if !ok {
		return
	}
	channel, err := h.channel.UpdateChannel(r.Context(), int(channelID), userID, body.Name, body.Description)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toChannel(channel))
}

func (h *Handler) DeleteChannel(w http.ResponseWriter, r *http.Request, channelID openapi.ChannelId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(channelID)) {
		badID(w)
		return
	}
	if err := h.channel.DeleteChannel(r.Context(), int(channelID), userID); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) SubscribeToChannel(w http.ResponseWriter, r *http.Request, channelID openapi.ChannelId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(channelID)) {
		badID(w)
		return
	}
	if err := h.subscription.Subscribe(r.Context(), userID, int(channelID)); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) UnsubscribeFromChannel(w http.ResponseWriter, r *http.Request, channelID openapi.ChannelId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(channelID)) {
		badID(w)
		return
	}
	if err := h.subscription.Unsubscribe(r.Context(), userID, int(channelID)); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) GetChannelSubscriptionStatus(w http.ResponseWriter, r *http.Request, channelID openapi.ChannelId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(channelID)) {
		badID(w)
		return
	}
	status, err := h.subscription.IsSubscribed(r.Context(), userID, int(channelID))
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, openapi.SubscriptionStatus{Subscribed: status})
}

func (h *Handler) GetChannelSubscribersCount(w http.ResponseWriter, r *http.Request, channelID openapi.ChannelId) {
	if !validID(int(channelID)) {
		badID(w)
		return
	}
	count, err := h.subscription.GetSubscribersCount(r.Context(), int(channelID))
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, openapi.CountResponse{Count: count})
}

func (h *Handler) ResetSubscriptionNewVideosCount(w http.ResponseWriter, r *http.Request, channelID openapi.ChannelId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(channelID)) {
		badID(w)
		return
	}
	if err := h.subscription.ResetNewVideosCount(r.Context(), userID, int(channelID)); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) ListMySubscriptions(w http.ResponseWriter, r *http.Request, params openapi.ListMySubscriptionsParams) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	limit, offset, err := pageParams(params.Limit, params.Offset)
	if err != nil {
		badRequest(w, "INVALID_PAGINATION", err.Error())
		return
	}
	page, err := h.subscription.GetUserSubscriptions(r.Context(), userID, limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	items := make([]openapi.Subscription, 0, len(page.Items))
	for _, item := range page.Items {
		if item != nil {
			items = append(items, toSubscription(item))
		}
	}
	response.RespondWithJSON(w, http.StatusOK, openapi.SubscriptionPage{
		Items: items, Limit: page.Limit, Offset: page.Offset, TotalCount: page.TotalCount,
	})
}
