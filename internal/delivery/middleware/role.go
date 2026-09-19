package middleware

import (
	"ZVideo/internal/delivery/response"
	"ZVideo/internal/domain"
	"log/slog"
	"net/http"
	"strings"
)

func RequireRole(allowedRoles ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := domain.GetLogger(r.Context())
			userCtx, err := GetUserFromContext(r.Context())
			if err != nil || userCtx == nil {
				logger.WarnContext(r.Context(), "Role check failed because user context is missing", slog.Any("error", err))
				response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
				return
			}
			for _, role := range allowedRoles {
				if strings.EqualFold(userCtx.Role, role) {
					next.ServeHTTP(w, r)
					return
				}
			}
			logger.WarnContext(r.Context(), "User does not have required role", slog.String("role", userCtx.Role))
			response.RespondWithError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		})
	}
}
