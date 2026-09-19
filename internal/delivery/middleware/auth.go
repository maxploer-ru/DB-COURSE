package middleware

import (
	"ZVideo/internal/delivery/response"
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"context"
	"net/http"
	"strings"
)

func Auth(authSvc service.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := domain.GetLogger(r.Context())
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			parts := strings.Fields(header)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
				logger.WarnContext(r.Context(), "Authorization header is missing or malformed")
				response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Bearer token is required")
				return
			}
			token := parts[1]
			accessTokenData, err := authSvc.ValidateAccessToken(r.Context(), token)
			if err != nil {
				response.HandleDomainError(w, err)
				return
			}
			userCtx := &UserContext{
				UserID: accessTokenData.UserID,
				Role:   accessTokenData.Role,
			}
			ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
			ctx = domain.WithUserID(ctx, accessTokenData.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
