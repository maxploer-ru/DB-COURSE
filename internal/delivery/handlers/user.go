package handlers

import (
	"ZVideo/internal/delivery/openapi"
	"ZVideo/internal/delivery/response"
	"net/http"
)

func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	user, err := h.user.GetProfile(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toUser(user))
}

func (h *Handler) DeleteCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if err := h.user.DeleteAccount(r.Context(), userID); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) UpdateCurrentUserNotificationSettings(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	body, ok := decodeJSON[openapi.UpdateCurrentUserNotificationSettingsJSONRequestBody](w, r)
	if !ok {
		return
	}
	if err := h.user.SetNotificationsSettings(r.Context(), userID, body.Enabled); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request, params openapi.ListUsersParams) {
	adminID, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	limit, offset, err := pageParams(params.Limit, params.Offset)
	if err != nil {
		badRequest(w, "INVALID_PAGINATION", err.Error())
		return
	}
	page, err := h.admin.ListUsers(r.Context(), adminID, limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	items := make([]openapi.User, 0, len(page.Items))
	for _, user := range page.Items {
		if user != nil {
			items = append(items, toUser(user))
		}
	}
	response.RespondWithJSON(w, http.StatusOK, openapi.UserPage{
		Items: items, Limit: page.Limit, Offset: page.Offset, TotalCount: page.TotalCount,
	})
}

func (h *Handler) UpdateUserStatus(w http.ResponseWriter, r *http.Request, userID openapi.UserId) {
	adminID, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	if !validID(int(userID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.UpdateUserStatusJSONRequestBody](w, r)
	if !ok {
		return
	}
	var err error
	switch body.Status {
	case openapi.Banned:
		err = h.admin.BanUser(r.Context(), adminID, int(userID))
	case openapi.Active:
		err = h.admin.UnbanUser(r.Context(), adminID, int(userID))
	default:
		badRequest(w, "INVALID_USER_STATUS", "Unsupported user status")
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	updated, err := h.user.GetProfile(r.Context(), int(userID))
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toUser(updated))
}

func (h *Handler) ReplaceUserRole(w http.ResponseWriter, r *http.Request, userID openapi.UserId) {
	adminID, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	if !validID(int(userID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.ReplaceUserRoleJSONRequestBody](w, r)
	if !ok {
		return
	}
	if err := h.admin.ChangeUserRole(r.Context(), adminID, int(userID), string(body.Role)); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) GetUserChannel(w http.ResponseWriter, r *http.Request, userID openapi.UserId) {
	if !validID(int(userID)) {
		badID(w)
		return
	}
	channel, err := h.channel.GetChannelByUserID(r.Context(), int(userID))
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toChannel(channel))
}
