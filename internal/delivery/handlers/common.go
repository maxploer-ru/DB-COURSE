package handlers

import (
	"ZVideo/internal/delivery/middleware"
	"ZVideo/internal/delivery/openapi"
	"ZVideo/internal/delivery/response"
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	openapiTypes "github.com/oapi-codegen/runtime/types"
)

const (
	defaultPageLimit = 20
	maxPageLimit     = 100
	maxJSONBodySize  = 1 << 20
	uploadURLTTL     = 15 * time.Minute
	streamURLTTL     = time.Hour
)

type Handler struct {
	auth               service.AuthService
	admin              service.AdminService
	user               service.UserService
	channel            service.ChannelService
	subscription       service.SubscriptionService
	video              service.VideoService
	videoInteraction   service.VideoInteractionService
	comment            service.CommentService
	commentInteraction service.CommentInteractionService
	playlist           service.PlaylistService
	community          service.CommunityService
}

func NewHandler(
	auth service.AuthService,
	admin service.AdminService,
	user service.UserService,
	channel service.ChannelService,
	subscription service.SubscriptionService,
	video service.VideoService,
	videoInteraction service.VideoInteractionService,
	comment service.CommentService,
	commentInteraction service.CommentInteractionService,
	playlist service.PlaylistService,
	community service.CommunityService,
) *Handler {
	return &Handler{
		auth:               auth,
		admin:              admin,
		user:               user,
		channel:            channel,
		subscription:       subscription,
		video:              video,
		videoInteraction:   videoInteraction,
		comment:            comment,
		commentInteraction: commentInteraction,
		playlist:           playlist,
		community:          community,
	}
}

var _ openapi.ServerInterface = (*Handler)(nil)

func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var value T
	if r.Body == nil {
		badRequest(w, "INVALID_REQUEST", "Request body is required")
		return value, false
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBodySize))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		badRequest(w, "INVALID_REQUEST", "Request body is invalid")
		return value, false
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		badRequest(w, "INVALID_REQUEST", "Request body must contain exactly one JSON value")
		return value, false
	}
	return value, true
}

func currentUser(r *http.Request) (int, string, bool) {
	user, err := middleware.GetUserFromContext(r.Context())
	if err != nil || user == nil || user.UserID < 1 {
		return 0, "", false
	}
	return user.UserID, user.Role, true
}

func requireCurrentUser(w http.ResponseWriter, r *http.Request) (int, string, bool) {
	userID, role, ok := currentUser(r)
	if !ok {
		response.RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authorized")
	}
	return userID, role, ok
}

func requireAdmin(w http.ResponseWriter, r *http.Request) (int, bool) {
	userID, role, ok := requireCurrentUser(w, r)
	if !ok {
		return 0, false
	}
	if !strings.EqualFold(role, domain.RoleAdmin) {
		response.RespondWithError(w, http.StatusForbidden, "FORBIDDEN", "You don't have permission to access this resource")
		return 0, false
	}
	return userID, true
}

func pageParams(limit, offset *int) (int, int, error) {
	pageLimit := defaultPageLimit
	if limit != nil {
		pageLimit = *limit
	}
	pageOffset := 0
	if offset != nil {
		pageOffset = *offset
	}
	if pageLimit < 1 || pageLimit > maxPageLimit {
		return 0, 0, fmt.Errorf("limit must be between 1 and %d", maxPageLimit)
	}
	if pageOffset < 0 {
		return 0, 0, errors.New("offset must be non-negative")
	}
	return pageLimit, pageOffset, nil
}

func validID(id int) bool {
	return id > 0
}

func badID(w http.ResponseWriter) {
	badRequest(w, "INVALID_PARAMETER", "Path identifiers must be positive integers")
}

func badRequest(w http.ResponseWriter, code, message string) {
	response.RespondWithError(w, http.StatusBadRequest, code, message)
}

func handleError(w http.ResponseWriter, err error) {
	response.HandleDomainError(w, err)
}

func writeNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func bearerToken(r *http.Request) string {
	parts := strings.Fields(strings.TrimSpace(r.Header.Get("Authorization")))
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

func toRole(role *domain.Role) *openapi.Role {
	if role == nil {
		return nil
	}
	return &openapi.Role{Id: role.ID, Name: openapi.RoleName(role.Name), IsDefault: role.IsDefault}
}

func toUser(user *domain.User) openapi.User {
	return openapi.User{
		Id:                   user.ID,
		Username:             user.Username,
		Email:                openapiTypes.Email(user.Email),
		IsActive:             user.IsActive,
		NotificationsEnabled: user.NotificationsEnabled,
		CreatedAt:            user.CreatedAt,
		UpdatedAt:            user.UpdatedAt,
		Role:                 toRole(user.Role),
	}
}

func toChannel(channel *domain.Channel) openapi.Channel {
	return openapi.Channel{
		Id:            channel.ID,
		UserId:        channel.UserID,
		OwnerUsername: channel.OwnerUsername,
		Name:          channel.Name,
		Description:   channel.Description,
		CreatedAt:     channel.CreatedAt,
	}
}

func toSubscription(subscription *domain.Subscription) openapi.Subscription {
	return openapi.Subscription{
		UserId:         subscription.UserID,
		ChannelId:      subscription.ChannelID,
		ChannelName:    subscription.ChannelName,
		NewVideosCount: subscription.NewVideosCount,
		SubscribedAt:   subscription.SubscribedAt,
	}
}

func toVideo(video *domain.Video) openapi.Video {
	result := openapi.Video{
		Id:               video.ID,
		ChannelId:        video.ChannelID,
		ChannelName:      video.ChannelName,
		Title:            video.Title,
		Description:      video.Description,
		Filepath:         video.Filepath,
		OriginalFilename: video.OriginalFilename,
		Status:           openapi.VideoStatus(video.Status),
		CreatedAt:        video.CreatedAt,
	}
	if video.Stats != nil {
		result.Stats = &openapi.VideoStats{
			Views:    video.Stats.Views,
			Likes:    video.Stats.Likes,
			Dislikes: video.Stats.Dislikes,
			Comments: video.Stats.Comments,
		}
	}
	return result
}

func toComment(comment *domain.Comment) openapi.Comment {
	return openapi.Comment{
		Id:        comment.ID,
		UserId:    comment.UserID,
		Username:  comment.Username,
		VideoId:   comment.VideoID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}
}

func toPlaylist(playlist *domain.Playlist) openapi.Playlist {
	return openapi.Playlist{
		Id:          playlist.ID,
		ChannelId:   playlist.ChannelID,
		Name:        playlist.Name,
		Description: playlist.Description,
		CreatedAt:   playlist.CreatedAt,
	}
}

func toPlaylistItem(item *domain.PlaylistItem) openapi.PlaylistItem {
	return openapi.PlaylistItem{
		PlaylistId:  item.PlaylistID,
		VideoId:     item.VideoID,
		Number:      item.Number,
		AddedAt:     item.AddedAt,
		VideoTitle:  item.VideoTitle,
		ChannelName: item.ChannelName,
		VideoStatus: openapi.VideoStatus(item.VideoStatus),
	}
}

func toCommunityPost(post *domain.CommunityPost) openapi.CommunityPost {
	return openapi.CommunityPost{
		Id:        post.ID,
		ChannelId: post.ChannelID,
		UserId:    post.UserID,
		Username:  post.Username,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
	}
}

func toCommunityComment(comment *domain.CommunityComment) openapi.CommunityComment {
	return openapi.CommunityComment{
		Id:        comment.ID,
		PostId:    comment.PostID,
		UserId:    comment.UserID,
		Username:  comment.Username,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}
}
