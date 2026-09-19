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
	GetChannelCommunity(ctx context.Context, channelID int, limit, offset int) (*domain.Community, error)
	GetMyCommunity(ctx context.Context, userID int, limit, offset int) (*domain.Community, error)
	GetPostComments(ctx context.Context, postID int, limit, offset int) (*domain.PageResponse[*domain.CommunityComment], error)
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
}

func NewCommunityService(communityRepo repository.CommunityRepository, channelSvc ChannelService) CommunityService {
	return &communityService{
		communityRepo: communityRepo,
		channelSvc:    channelSvc,
	}
}

func (s *communityService) GetChannelCommunity(ctx context.Context, channelID int, limit, offset int) (*domain.Community, error) {
	channel, err := s.channelSvc.GetChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}

	posts, err := s.communityRepo.ListPostsByChannel(ctx, channelID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list community posts: %w", err)
	}
	total, err := s.communityRepo.CountPostsByChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("count community posts: %w", err)
	}

	return &domain.Community{
		Channel:    channel,
		Posts:      posts,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func (s *communityService) GetMyCommunity(ctx context.Context, userID int, limit, offset int) (*domain.Community, error) {
	channel, err := s.channelSvc.GetChannelByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get channel by user: %w", err)
	}
	if channel == nil {
		return nil, domain.ErrChannelNotFound
	}
	return s.GetChannelCommunity(ctx, channel.ID, limit, offset)
}

func (s *communityService) GetPostComments(ctx context.Context, postID int, limit, offset int) (*domain.PageResponse[*domain.CommunityComment], error) {
	post, err := s.communityRepo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("get community post: %w", err)
	}
	if post == nil {
		return nil, domain.ErrCommunityPostNotFound
	}

	comments, err := s.communityRepo.ListCommentsByPost(ctx, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	total, err := s.communityRepo.CountCommentsByPost(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("count comments: %w", err)
	}
	return &domain.PageResponse[*domain.CommunityComment]{
		Items:      comments,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func (s *communityService) CreatePost(ctx context.Context, channelID, userID int, content string) (*domain.CommunityPost, error) {
	isOwner, err := s.channelSvc.IsOwner(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("check channel owner: %w", err)
	}
	if !isOwner {
		return nil, domain.ErrForbidden
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, domain.ErrCommunityPostContentEmpty
	}

	post := &domain.CommunityPost{
		ChannelID: channelID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if err := s.communityRepo.CreatePost(ctx, post); err != nil {
		return nil, fmt.Errorf("create community post: %w", err)
	}

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
