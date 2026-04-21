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
	"strings"
	"time"
)

const refreshCookieName = "refresh_token"

type AuthHandler struct {
	authSvc service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "Register"))

	req := dto.RegisterRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Failed to decode registration request",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	logger = logger.With(
		slog.String("username", req.Username),
		slog.String("email", req.Email))

	err := h.authSvc.Register(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		logger.WarnContext(r.Context(), "User registration failed",
			slog.String("error", err.Error()),
		)
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "Login"))

	req := dto.LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WarnContext(r.Context(), "Failed to decode login request",
			slog.String("error", err.Error()))

		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	logger = logger.With(
		slog.String("email", req.Email))

	result, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		logger.WarnContext(r.Context(), "User login failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	setRefreshCookie(w, r, result.RefreshToken, result.RefreshExpiresAt)
	resp := mappers.ToAuthResponse(result.User, result.AccessToken, "")
	response.RespondWithJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "Refresh"))

	refreshToken, err := readRefreshCookie(r)
	if err != nil {
		logger.WarnContext(r.Context(), "Refresh cookie is missing", slog.String("error", err.Error()))
		response.HandleDomainError(w, domain.ErrInvalidRefreshToken)
		return
	}

	result, err := h.authSvc.Refresh(r.Context(), refreshToken)
	if err != nil {
		logger.WarnContext(r.Context(), "User refresh failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	setRefreshCookie(w, r, result.RefreshToken, result.RefreshExpiresAt)
	resp := mappers.ToAuthResponse(result.User, result.AccessToken, "")
	response.RespondWithJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "Logout"))

	accessToken := ""
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken = authHeader[7:]
	}

	refreshToken, _ := readRefreshCookie(r)
	err := h.authSvc.Logout(r.Context(), accessToken, refreshToken)
	if err != nil {
		logger.WarnContext(r.Context(), "User logout failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	clearRefreshCookie(w, r)
	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func (h *AuthHandler) ValidateAccessToken(w http.ResponseWriter, r *http.Request) {
	logger := domain.GetLogger(r.Context()).With(
		slog.String("handler", "ValidateAccessToken"))

	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		logger.WarnContext(r.Context(), "Invalid authorization header",
			slog.String("error", "Invalid authorization header"))

		response.RespondWithError(w, http.StatusUnauthorized, "missing_token", "Authorization header with Bearer token required")
		return
	}
	token := authHeader[7:]

	tokenData, err := h.authSvc.ValidateAccessToken(r.Context(), token)
	if err != nil {
		logger.WarnContext(r.Context(), "User validate access token failed",
			slog.String("error", err.Error()))

		response.HandleDomainError(w, err)
		return
	}

	resp := mappers.ToValidateTokenResponse(tokenData)
	response.RespondWithJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	user, err := h.authSvc.GetMe(r.Context(), userCtx.UserID)
	if err != nil {
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, mappers.ToUserResponse(user))
}

func (h *AuthHandler) UpdateNotificationsSettings(w http.ResponseWriter, r *http.Request) {
	userCtx, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
		return
	}

	var req dto.NotificationsSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if req.Enabled == nil {
		response.HandleDomainError(w, domain.ErrInvalidNotificationSettings)
		return
	}

	user, err := h.authSvc.SetNotificationsEnabled(r.Context(), userCtx.UserID, *req.Enabled)
	if err != nil {
		response.HandleDomainError(w, err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, mappers.ToUserResponse(user))
}

func setRefreshCookie(w http.ResponseWriter, r *http.Request, token string, expiresAt time.Time) {
	secure := isSecureRequest(r)
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})
}

func clearRefreshCookie(w http.ResponseWriter, r *http.Request) {
	secure := isSecureRequest(r)
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func readRefreshCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		return "", err
	}
	if cookie.Value == "" {
		return "", http.ErrNoCookie
	}
	return cookie.Value, nil
}
