package response

import (
	"ZVideo/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RespondWithError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func RespondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func HandleDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, context.Canceled):
		RespondWithError(w, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "The request was canceled by the client")

	case errors.Is(err, domain.ErrInvalidUsername):
		RespondWithError(w, http.StatusBadRequest, "INVALID_USERNAME", "The username is invalid")
	case errors.Is(err, domain.ErrInvalidUserEmail):
		RespondWithError(w, http.StatusBadRequest, "INVALID_USER_EMAIL", "The email is invalid")
	case errors.Is(err, domain.ErrWeakPassword):
		RespondWithError(w, http.StatusBadRequest, "WEAK_PASSWORD", "The password is too weak")

	case errors.Is(err, domain.ErrInvalidUserCredentials):
		RespondWithError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
	case errors.Is(err, domain.ErrUserIsBanned):
		RespondWithError(w, http.StatusForbidden, "USER_BANNED", "User account is banned")
	case errors.Is(err, domain.ErrUserNotFound):
		RespondWithError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	case errors.Is(err, domain.ErrInvalidAccessToken), errors.Is(err, domain.ErrInvalidRefreshToken):
		RespondWithError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token")
	case errors.Is(err, domain.ErrForbidden):
		RespondWithError(w, http.StatusForbidden, "FORBIDDEN", "You don't have permission to access this resource")

	case errors.Is(err, domain.ErrUserNameAlreadyExists):
		RespondWithError(w, http.StatusConflict, "USERNAME_TAKEN", "Username already taken")
	case errors.Is(err, domain.ErrUserEmailAlreadyExists):
		RespondWithError(w, http.StatusConflict, "EMAIL_TAKEN", "Email already registered")

	case errors.Is(err, domain.ErrChannelNotFound):
		RespondWithError(w, http.StatusNotFound, "CHANNEL_NOT_FOUND", "Channel not found")
	case errors.Is(err, domain.ErrChannelAlreadyExists):
		RespondWithError(w, http.StatusConflict, "CHANNEL_ALREADY_EXISTS", "Channel already exists")
	case errors.Is(err, domain.ErrChannelNameAlreadyExists):
		RespondWithError(w, http.StatusConflict, "CHANNEL_NAME_TAKEN", "Channel name already exists")

	case errors.Is(err, domain.ErrVideoNotFound):
		RespondWithError(w, http.StatusNotFound, "VIDEO_NOT_FOUND", "Video not found")
	case errors.Is(err, domain.ErrPlaylistNotFound):
		RespondWithError(w, http.StatusNotFound, "PLAYLIST_NOT_FOUND", "Playlist not found")
	case errors.Is(err, domain.ErrPlaylistNameEmpty):
		RespondWithError(w, http.StatusBadRequest, "PLAYLIST_NAME_EMPTY", "Playlist name cannot be empty")
	case errors.Is(err, domain.ErrPlaylistVideoChannelMismatch):
		RespondWithError(w, http.StatusBadRequest, "PLAYLIST_VIDEO_CHANNEL_MISMATCH", "Video belongs to another channel")

	case errors.Is(err, domain.ErrSelfSubscription):
		RespondWithError(w, http.StatusBadRequest, "SUBSCRIPTION_NOT_FOUND", "Cannot subscribe to your own channel")

	case errors.Is(err, domain.ErrAlreadyRated):
		RespondWithError(w, http.StatusConflict, "ALREADY_RATED", "This video has already been rated")
	case errors.Is(err, domain.ErrRatingNotFound):
		RespondWithError(w, http.StatusNotFound, "RATING_NOT_FOUND", "Rating not found")

	case errors.Is(err, domain.ErrCommentNotFound):
		RespondWithError(w, http.StatusNotFound, "COMMENT_NOT_FOUND", "Comment not found")

	case errors.Is(err, domain.ErrCommentRatingNotFound):
		RespondWithError(w, http.StatusNotFound, "RATING_NOT_FOUND", "Comment rating not found")
	case errors.Is(err, domain.ErrInvalidNotificationSettings):
		RespondWithError(w, http.StatusBadRequest, "INVALID_NOTIFICATION_SETTINGS", "Notification settings payload is invalid")

	case errors.Is(err, domain.ErrInternalServer):
		RespondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	default:
		RespondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong")
	}
}
