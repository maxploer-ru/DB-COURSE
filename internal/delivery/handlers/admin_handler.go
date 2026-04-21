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

type AdminHandler struct {
	adminSvc service.AdminService
}

func NewAdminHandler(adminSvc service.AdminService) *AdminHandler {
	return &AdminHandler{adminSvc: adminSvc}
}

func (h *AdminHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "BanUser"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to BanUser", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}
	logger = logger.With(slog.Int("admin_id", userCtx.UserID))

	targetID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Invalid target user ID", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid user id")
		return
	}
	logger = logger.With(slog.Int("target_user_id", targetID))

	if err := h.adminSvc.BanUser(r.Context(), targetID); err != nil {
		logger.WarnContext(r.Context(), "Failed to ban user", slog.String("error", err.Error()))
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User banned successfully"})
}

func (h *AdminHandler) UnbanUser(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(slog.String("handler", "UnbanUser"))

	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		logger.WarnContext(r.Context(), "Unauthorized access to UnbanUser", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}
	logger = logger.With(slog.Int("admin_id", userCtx.UserID))

	targetID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logger.WarnContext(r.Context(), "Invalid target user ID", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid user id")
		return
	}
	logger = logger.With(slog.Int("target_user_id", targetID))

	if err := h.adminSvc.UnbanUser(r.Context(), targetID); err != nil {
		logger.WarnContext(r.Context(), "Failed to unban user", slog.String("error", err.Error()))
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User unbanned successfully"})
}
