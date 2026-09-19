package handlers

import (
	"ZVideo/internal/delivery/openapi"
	"ZVideo/internal/delivery/response"
	"net/http"
)

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeJSON[openapi.RegisterUserJSONRequestBody](w, r)
	if !ok {
		return
	}
	user, err := h.auth.Register(r.Context(), body.Username, string(body.Email), body.Password)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusCreated, toUser(user))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeJSON[openapi.LoginJSONRequestBody](w, r)
	if !ok {
		return
	}
	result, err := h.auth.Login(r.Context(), string(body.Email), body.Password)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, openapi.AuthResponse{
		User:             toUser(result.User),
		AccessToken:      result.AccessToken,
		RefreshToken:     result.RefreshToken,
		RefreshExpiresAt: result.RefreshExpiresAt,
	})
}

func (h *Handler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeJSON[openapi.RefreshTokensJSONRequestBody](w, r)
	if !ok {
		return
	}
	result, err := h.auth.Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, openapi.AuthResponse{
		User:             toUser(result.User),
		AccessToken:      result.AccessToken,
		RefreshToken:     result.RefreshToken,
		RefreshExpiresAt: result.RefreshExpiresAt,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeJSON[openapi.LogoutJSONRequestBody](w, r)
	if !ok {
		return
	}
	if err := h.auth.Logout(r.Context(), bearerToken(r), body.RefreshToken); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}
