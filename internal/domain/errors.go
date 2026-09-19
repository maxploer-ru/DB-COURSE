package domain

import "errors"

var (
	ErrInvalidUsername        = errors.New("invalid username")
	ErrInvalidUserEmail       = errors.New("invalid user email")
	ErrWeakPassword           = errors.New("weak password")
	ErrUserNameAlreadyExists  = errors.New("username already exists")
	ErrUserEmailAlreadyExists = errors.New("user email already exists")
	ErrInvalidUserCredentials = errors.New("invalid user credentials")
	ErrUserIsBanned           = errors.New("user is banned")
	ErrUserNotFound           = errors.New("user not found")
	ErrInvalidAccessToken     = errors.New("invalid access token")
	ErrInvalidRefreshToken    = errors.New("invalid refresh token")
	ErrForbidden              = errors.New("forbidden")
	ErrRoleNotFound           = errors.New("role not found")
)

var (
	ErrChannelNotFound             = errors.New("channel not found")
	ErrChannelAlreadyExists        = errors.New("channel already exists")
	ErrChannelNameAlreadyExists    = errors.New("channel name already exists")
	ErrSelfSubscription            = errors.New("self subscription")
	ErrInvalidNotificationSettings = errors.New("invalid notification settings")
	ErrInvalidChannelName          = errors.New("invalid channel name")
)

var (
	ErrVideoNotFound                = errors.New("video not found")
	ErrPlaylistNotFound             = errors.New("playlist not found")
	ErrPlaylistNameEmpty            = errors.New("playlist name cannot be empty")
	ErrPlaylistVideoChannelMismatch = errors.New("video belongs to another channel")
	ErrVideoNotReady                = errors.New("video is not ready yet")
)

var (
	ErrAlreadyRated   = errors.New("already rated")
	ErrRatingNotFound = errors.New("rating not found")
)

var (
	ErrCommentNotFound              = errors.New("comment not found")
	ErrInvalidCommentContent        = errors.New("comment content cannot be empty")
	ErrCommentRatingNotFound        = errors.New("comment rating not found")
	ErrCommunityPostNotFound        = errors.New("community post not found")
	ErrCommunityCommentNotFound     = errors.New("community comment not found")
	ErrCommunityPostContentEmpty    = errors.New("community post content cannot be empty")
	ErrCommunityCommentContentEmpty = errors.New("community comment content cannot be empty")
	ErrInvalidRatingAction          = errors.New("invalid rating action")
)

var (
	ErrInternalServer = errors.New("internal server error")
)
