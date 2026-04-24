package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type CommunityService interface {
	GetChannelCommunity(ctx context.Context, channelID int) (*domain.Community, error)
	GetMyCommunity(ctx context.Context, userID int) (*domain.Community, error)
	CreatePost(ctx context.Context, channelID, userID int, content string) (*domain.CommunityPost, error)
	UpdatePost(ctx context.Context, postID, userID int, content string) (*domain.CommunityPost, error)
	DeletePost(ctx context.Context, postID, userID int) error
	CreateComment(ctx context.Context, postID, userID int, content string) (*domain.CommunityComment, error)
	UpdateComment(ctx context.Context, commentID, userID int, content string) (*domain.CommunityComment, error)
	DeleteComment(ctx context.Context, commentID, userID int) error
}

type communityService struct {
	communityRepo repository.CommunityRepository
	channelSvc    ChannelService
	userRepo      repository.UserRepository
}

func NewCommunityService(communityRepo repository.CommunityRepository, channelSvc ChannelService, userRepo repository.UserRepository) CommunityService {
	return &communityService{communityRepo: communityRepo, channelSvc: channelSvc, userRepo: userRepo}
}

func (s *communityService) resolveUsername(ctx context.Context, userID int, usernameCache map[int]string) string {
	if usernameCache == nil {
		usernameCache = map[int]string{}
	}
	if username, ok := usernameCache[userID]; ok {
		return username
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil || user.Username == "" {
		username := fmt.Sprintf("Пользователь #%d", userID)
		usernameCache[userID] = username
		return username
	}

	usernameCache[userID] = user.Username
	return user.Username
}

func (s *communityService) enrichPostAuthor(ctx context.Context, post *domain.CommunityPost, usernameCache map[int]string) {
	if post == nil {
		return
	}
	post.Username = s.resolveUsername(ctx, post.UserID, usernameCache)
}

func (s *communityService) enrichCommentAuthor(ctx context.Context, comment *domain.CommunityComment, usernameCache map[int]string) {
	if comment == nil {
		return
	}
	comment.Username = s.resolveUsername(ctx, comment.UserID, usernameCache)
}

func (s *communityService) GetChannelCommunity(ctx context.Context, channelID int) (*domain.Community, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommunityService"),
		slog.String("operation", "GetChannelCommunity"),
		slog.Int("channel_id", channelID),
	)

	channel, err := s.channelSvc.GetChannel(ctx, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get channel", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get channel: %w", err)
	}

	posts, err := s.communityRepo.ListPostsByChannel(ctx, channelID, 100, 0)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list community posts", slog.String("error", err.Error()))
		return nil, fmt.Errorf("list community posts: %w", err)
	}

	community := &domain.Community{Channel: channel, Posts: make([]*domain.CommunityPostWithComments, 0, len(posts))}
	usernameCache := make(map[int]string)
	for _, post := range posts {
		comments, err := s.communityRepo.ListCommentsByPost(ctx, post.ID, 100, 0)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to list comments for post", slog.Int("post_id", post.ID), slog.String("error", err.Error()))
			return nil, fmt.Errorf("list community comments: %w", err)
		}
		s.enrichPostAuthor(ctx, post, usernameCache)
		for _, comment := range comments {
			s.enrichCommentAuthor(ctx, comment, usernameCache)
		}
		community.Posts = append(community.Posts, &domain.CommunityPostWithComments{Post: post, Comments: comments})
	}

	return community, nil
}

func (s *communityService) GetMyCommunity(ctx context.Context, userID int) (*domain.Community, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommunityService"),
		slog.String("operation", "GetMyCommunity"),
		slog.Int("user_id", userID),
	)

	channel, err := s.channelSvc.GetChannelByUserID(ctx, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get channel by user", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get channel by user: %w", err)
	}

	return s.GetChannelCommunity(ctx, channel.ID)
}

func (s *communityService) CreatePost(ctx context.Context, channelID, userID int, content string) (*domain.CommunityPost, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommunityService"),
		slog.String("operation", "CreatePost"),
		slog.Int("channel_id", channelID),
		slog.Int("user_id", userID),
	)

	isOwner, err := s.channelSvc.IsOwner(ctx, channelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel ownership", slog.String("error", err.Error()))
		return nil, fmt.Errorf("check channel owner: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not the channel owner")
		return nil, domain.ErrForbidden
	}

	content = strings.TrimSpace(content)
	if content == "" {
		logger.WarnContext(ctx, "Community post content is empty")
		return nil, domain.ErrCommunityPostContentEmpty
	}

	post := &domain.CommunityPost{
		ChannelID: channelID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if err := s.communityRepo.CreatePost(ctx, post); err != nil {
		logger.ErrorContext(ctx, "Failed to create community post", slog.String("error", err.Error()))
		return nil, fmt.Errorf("create community post: %w", err)
	}
	s.enrichPostAuthor(ctx, post, map[int]string{})

	logger.InfoContext(ctx, "Community post created successfully", slog.Int("post_id", post.ID))
	return post, nil
}

func (s *communityService) UpdatePost(ctx context.Context, postID, userID int, content string) (*domain.CommunityPost, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommunityService"),
		slog.String("operation", "UpdatePost"),
		slog.Int("post_id", postID),
		slog.Int("user_id", userID),
	)

	post, err := s.communityRepo.GetPostByID(ctx, postID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get community post", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get community post: %w", err)
	}
	if post == nil {
		logger.WarnContext(ctx, "Community post not found")
		return nil, domain.ErrCommunityPostNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, post.ChannelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel ownership", slog.String("error", err.Error()))
		return nil, fmt.Errorf("check channel owner: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not the channel owner")
		return nil, domain.ErrForbidden
	}

	content = strings.TrimSpace(content)
	if content == "" {
		logger.WarnContext(ctx, "Community post content is empty")
		return nil, domain.ErrCommunityPostContentEmpty
	}

	post.Content = content
	if err := s.communityRepo.UpdatePost(ctx, post); err != nil {
		logger.ErrorContext(ctx, "Failed to update community post", slog.String("error", err.Error()))
		return nil, fmt.Errorf("update community post: %w", err)
	}
	s.enrichPostAuthor(ctx, post, map[int]string{})

	logger.InfoContext(ctx, "Community post updated successfully")
	return post, nil
}

func (s *communityService) DeletePost(ctx context.Context, postID, userID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommunityService"),
		slog.String("operation", "DeletePost"),
		slog.Int("post_id", postID),
		slog.Int("user_id", userID),
	)

	post, err := s.communityRepo.GetPostByID(ctx, postID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get community post", slog.String("error", err.Error()))
		return fmt.Errorf("get community post: %w", err)
	}
	if post == nil {
		logger.WarnContext(ctx, "Community post not found")
		return domain.ErrCommunityPostNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, post.ChannelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel ownership", slog.String("error", err.Error()))
		return fmt.Errorf("check channel owner: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not the channel owner")
		return domain.ErrForbidden
	}

	if err := s.communityRepo.DeletePost(ctx, postID); err != nil {
		logger.ErrorContext(ctx, "Failed to delete community post", slog.String("error", err.Error()))
		return fmt.Errorf("delete community post: %w", err)
	}

	logger.InfoContext(ctx, "Community post deleted successfully")
	return nil
}

func (s *communityService) CreateComment(ctx context.Context, postID, userID int, content string) (*domain.CommunityComment, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommunityService"),
		slog.String("operation", "CreateComment"),
		slog.Int("post_id", postID),
		slog.Int("user_id", userID),
	)

	post, err := s.communityRepo.GetPostByID(ctx, postID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get community post", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get community post: %w", err)
	}
	if post == nil {
		logger.WarnContext(ctx, "Community post not found")
		return nil, domain.ErrCommunityPostNotFound
	}

	content = strings.TrimSpace(content)
	if content == "" {
		logger.WarnContext(ctx, "Community comment content is empty")
		return nil, domain.ErrCommunityCommentContentEmpty
	}

	comment := &domain.CommunityComment{
		PostID:    postID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if err := s.communityRepo.CreateComment(ctx, comment); err != nil {
		logger.ErrorContext(ctx, "Failed to create community comment", slog.String("error", err.Error()))
		return nil, fmt.Errorf("create community comment: %w", err)
	}
	s.enrichCommentAuthor(ctx, comment, map[int]string{})

	logger.InfoContext(ctx, "Community comment created successfully", slog.Int("comment_id", comment.ID))
	return comment, nil
}

func (s *communityService) UpdateComment(ctx context.Context, commentID, userID int, content string) (*domain.CommunityComment, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommunityService"),
		slog.String("operation", "UpdateComment"),
		slog.Int("comment_id", commentID),
		slog.Int("user_id", userID),
	)

	comment, err := s.communityRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get community comment", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get community comment: %w", err)
	}
	if comment == nil {
		logger.WarnContext(ctx, "Community comment not found")
		return nil, domain.ErrCommunityCommentNotFound
	}

	if comment.UserID != userID {
		logger.WarnContext(ctx, "User is not the comment author")
		return nil, domain.ErrForbidden
	}

	content = strings.TrimSpace(content)
	if content == "" {
		logger.WarnContext(ctx, "Community comment content is empty")
		return nil, domain.ErrCommunityCommentContentEmpty
	}

	comment.Content = content
	if err := s.communityRepo.UpdateComment(ctx, comment); err != nil {
		logger.ErrorContext(ctx, "Failed to update community comment", slog.String("error", err.Error()))
		return nil, fmt.Errorf("update community comment: %w", err)
	}
	s.enrichCommentAuthor(ctx, comment, map[int]string{})

	logger.InfoContext(ctx, "Community comment updated successfully")
	return comment, nil
}

func (s *communityService) DeleteComment(ctx context.Context, commentID, userID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommunityService"),
		slog.String("operation", "DeleteComment"),
		slog.Int("comment_id", commentID),
		slog.Int("user_id", userID),
	)

	comment, err := s.communityRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get community comment", slog.String("error", err.Error()))
		return fmt.Errorf("get community comment: %w", err)
	}
	if comment == nil {
		logger.WarnContext(ctx, "Community comment not found")
		return domain.ErrCommunityCommentNotFound
	}

	if comment.UserID != userID {
		logger.WarnContext(ctx, "User is not the comment author")
		return domain.ErrForbidden
	}

	if err := s.communityRepo.DeleteComment(ctx, commentID); err != nil {
		logger.ErrorContext(ctx, "Failed to delete community comment", slog.String("error", err.Error()))
		return fmt.Errorf("delete community comment: %w", err)
	}

	logger.InfoContext(ctx, "Community comment deleted successfully")
	return nil
}
