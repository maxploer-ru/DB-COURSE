package middleware

import (
	"ZVideo/internal/delivery/response"
	"net/http"
)

func RequireRole(allowedRoles ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, err := GetUserFromContext(r.Context())
			if err != nil || userCtx == nil {
				response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
				return
			}
			for _, role := range allowedRoles {
				if userCtx.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			response.RespondWithError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		})
	}
}
