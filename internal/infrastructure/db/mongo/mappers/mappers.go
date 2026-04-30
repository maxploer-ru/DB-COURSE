package mappers

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/mongo/models"
)

func ToDomainRole(role *models.Role) *domain.Role {
	if role == nil {
		return nil
	}
	return &domain.Role{
		ID:        role.ID,
		Name:      role.Name,
		IsDefault: role.IsDefault,
	}
}

func FromDomainRole(role *domain.Role) *models.Role {
	if role == nil {
		return nil
	}
	return &models.Role{
		ID:        role.ID,
		Name:      role.Name,
		IsDefault: role.IsDefault,
	}
}

func ToDomainUser(user *models.User, role *models.Role) *domain.User {
	if user == nil {
		return nil
	}
	return &domain.User{
		ID:                   user.ID,
		Username:             user.Username,
		Email:                user.Email,
		PasswordHash:         user.PasswordHash,
		IsActive:             user.IsActive,
		NotificationsEnabled: user.NotificationsEnabled,
		Role:                 ToDomainRole(role),
	}
}

func FromDomainUser(user *domain.User) *models.User {
	if user == nil {
		return nil
	}
	roleID := 0
	if user.Role != nil {
		roleID = user.Role.ID
	}
	return &models.User{
		ID:                   user.ID,
		Username:             user.Username,
		Email:                user.Email,
		PasswordHash:         user.PasswordHash,
		IsActive:             user.IsActive,
		NotificationsEnabled: user.NotificationsEnabled,
		RoleID:               roleID,
	}
}

func ToDomainChannel(channel *models.Channel) *domain.Channel {
	if channel == nil {
		return nil
	}
	return &domain.Channel{
		ID:          channel.ID,
		UserID:      channel.UserID,
		Name:        channel.Name,
		Description: channel.Description,
		CreatedAt:   channel.CreatedAt,
	}
}

func FromDomainChannel(channel *domain.Channel) *models.Channel {
	if channel == nil {
		return nil
	}
	return &models.Channel{
		ID:          channel.ID,
		UserID:      channel.UserID,
		Name:        channel.Name,
		Description: channel.Description,
		CreatedAt:   channel.CreatedAt,
	}
}

func ToDomainVideo(video *models.Video, channelName string) *domain.Video {
	if video == nil {
		return nil
	}
	return &domain.Video{
		ID:          video.ID,
		ChannelID:   video.ChannelID,
		ChannelName: channelName,
		Title:       video.Title,
		Description: video.Description,
		Filepath:    video.Filepath,
		CreatedAt:   video.CreatedAt,
	}
}

func FromDomainVideo(video *domain.Video) *models.Video {
	if video == nil {
		return nil
	}
	return &models.Video{
		ID:          video.ID,
		ChannelID:   video.ChannelID,
		Title:       video.Title,
		Description: video.Description,
		Filepath:    video.Filepath,
		CreatedAt:   video.CreatedAt,
	}
}

func ToDomainComment(comment *models.Comment, username string) *domain.Comment {
	if comment == nil {
		return nil
	}
	return &domain.Comment{
		ID:        comment.ID,
		UserID:    comment.UserID,
		Username:  username,
		VideoID:   comment.VideoID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}
}

func FromDomainComment(comment *domain.Comment) *models.Comment {
	if comment == nil {
		return nil
	}
	return &models.Comment{
		ID:        comment.ID,
		UserID:    comment.UserID,
		VideoID:   comment.VideoID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}
}

func ToDomainCommentRating(rating *models.CommentRating) *domain.CommentRating {
	if rating == nil {
		return nil
	}
	return &domain.CommentRating{
		UserID:    rating.UserID,
		CommentID: rating.CommentID,
		Liked:     rating.Liked,
		RatedAt:   rating.RatedAt,
	}
}

func FromDomainCommentRating(rating *domain.CommentRating, id string) *models.CommentRating {
	if rating == nil {
		return nil
	}
	return &models.CommentRating{
		ID:        id,
		UserID:    rating.UserID,
		CommentID: rating.CommentID,
		Liked:     rating.Liked,
		RatedAt:   rating.RatedAt,
	}
}

func ToDomainVideoRating(rating *models.VideoRating) *domain.VideoRating {
	if rating == nil {
		return nil
	}
	return &domain.VideoRating{
		UserID:  rating.UserID,
		VideoID: rating.VideoID,
		Liked:   rating.Liked,
		RatedAt: rating.RatedAt,
	}
}

func FromDomainVideoRating(rating *domain.VideoRating, id string) *models.VideoRating {
	if rating == nil {
		return nil
	}
	return &models.VideoRating{
		ID:      id,
		UserID:  rating.UserID,
		VideoID: rating.VideoID,
		Liked:   rating.Liked,
		RatedAt: rating.RatedAt,
	}
}

func ToDomainSubscription(sub *models.Subscription) *domain.Subscription {
	if sub == nil {
		return nil
	}
	return &domain.Subscription{
		UserID:         sub.UserID,
		ChannelID:      sub.ChannelID,
		NewVideosCount: sub.NewVideosCount,
		SubscribedAt:   sub.SubscribedAt,
	}
}

func FromDomainSubscription(sub *domain.Subscription, id string) *models.Subscription {
	if sub == nil {
		return nil
	}
	return &models.Subscription{
		ID:             id,
		UserID:         sub.UserID,
		ChannelID:      sub.ChannelID,
		NewVideosCount: sub.NewVideosCount,
		SubscribedAt:   sub.SubscribedAt,
	}
}

func ToDomainViewing(viewing *models.Viewing) *domain.Viewing {
	if viewing == nil {
		return nil
	}
	return &domain.Viewing{
		ID:        viewing.ID,
		UserID:    viewing.UserID,
		VideoID:   viewing.VideoID,
		WatchedAt: viewing.WatchedAt,
	}
}

func FromDomainViewing(viewing *domain.Viewing) *models.Viewing {
	if viewing == nil {
		return nil
	}
	return &models.Viewing{
		ID:        viewing.ID,
		UserID:    viewing.UserID,
		VideoID:   viewing.VideoID,
		WatchedAt: viewing.WatchedAt,
	}
}

func ToDomainCommunityPost(post *models.CommunityPost) *domain.CommunityPost {
	if post == nil {
		return nil
	}
	return &domain.CommunityPost{
		ID:        post.ID,
		ChannelID: post.ChannelID,
		UserID:    post.UserID,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
	}
}

func FromDomainCommunityPost(post *domain.CommunityPost) *models.CommunityPost {
	if post == nil {
		return nil
	}
	return &models.CommunityPost{
		ID:        post.ID,
		ChannelID: post.ChannelID,
		UserID:    post.UserID,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
	}
}

func ToDomainCommunityComment(comment *models.CommunityComment) *domain.CommunityComment {
	if comment == nil {
		return nil
	}
	return &domain.CommunityComment{
		ID:        comment.ID,
		PostID:    comment.PostID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}
}

func FromDomainCommunityComment(comment *domain.CommunityComment) *models.CommunityComment {
	if comment == nil {
		return nil
	}
	return &models.CommunityComment{
		ID:        comment.ID,
		PostID:    comment.PostID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}
}

func ToDomainPlaylist(playlist *models.Playlist, items []domain.PlaylistItem) *domain.Playlist {
	if playlist == nil {
		return nil
	}
	return &domain.Playlist{
		ID:          playlist.ID,
		ChannelID:   playlist.ChannelID,
		Name:        playlist.Name,
		Description: playlist.Description,
		CreatedAt:   playlist.CreatedAt,
		Items:       items,
	}
}

func FromDomainPlaylist(playlist *domain.Playlist, items []models.PlaylistItem) *models.Playlist {
	if playlist == nil {
		return nil
	}
	return &models.Playlist{
		ID:          playlist.ID,
		ChannelID:   playlist.ChannelID,
		Name:        playlist.Name,
		Description: playlist.Description,
		CreatedAt:   playlist.CreatedAt,
		Items:       items,
	}
}
