package domain

import "errors"

var (
	ErrInvalidUsername  = errors.New("invalid username")
	ErrInvalidUserEmail = errors.New("invalid user email")
	ErrWeakPassword     = errors.New("weak password")

	ErrUserNameAlreadyExists  = errors.New("username already exists")
	ErrUserEmailAlreadyExists = errors.New("user email already exists")
	ErrInvalidUserCredentials = errors.New("invalid user credentials")
	ErrUserIsBanned           = errors.New("user is banned")
	ErrUserNotFound           = errors.New("user not found")

	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")

	ErrInternalServer = errors.New("internal server error")

	ErrChannelNotFound          = errors.New("channel not found")
	ErrChannelAlreadyExists     = errors.New("channel already exists")
	ErrChannelNameAlreadyExists = errors.New("channel name already exists")

	ErrSelfSubscription = errors.New("self subscription")

	ErrVideoNotFound                = errors.New("video not found")
	ErrPlaylistNotFound             = errors.New("playlist not found")
	ErrPlaylistNameEmpty            = errors.New("playlist name cannot be empty")
	ErrPlaylistVideoChannelMismatch = errors.New("video belongs to another channel")

	ErrAlreadyRated   = errors.New("already rated")
	ErrRatingNotFound = errors.New("rating not found")

	ErrCommentNotFound             = errors.New("comment not found")
	ErrCommentRatingNotFound       = errors.New("comment rating not found")
	ErrInvalidNotificationSettings = errors.New("invalid notification settings")

	ErrForbidden = errors.New("forbidden")
)
