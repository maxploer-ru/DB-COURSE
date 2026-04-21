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
			token := r.Header.Get("Authorization")
			if strings.HasPrefix(token, "Bearer ") {
				token = strings.TrimPrefix(token, "Bearer ")
			}
			accessTokenData, err := authSvc.ValidateAccessToken(r.Context(), token)
			if err != nil {
				response.RespondWithError(w, http.StatusUnauthorized, "INVALID_ACCESS_TOKEN", err.Error())
				return
			}
			userCtx := &UserContext{
				UserID: accessTokenData.UserID,
				Role:   accessTokenData.Role,
			}
			ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
			ctx = context.WithValue(ctx, domain.UserIDKey, accessTokenData.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
