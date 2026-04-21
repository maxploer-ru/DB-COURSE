package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type CommentService interface {
	Create(ctx context.Context, userID, videoID int, content string) (*domain.Comment, error)
	GetByID(ctx context.Context, id int) (*domain.Comment, error)
	ListByVideo(ctx context.Context, videoID int, limit, offset int) ([]*domain.Comment, error)
	Update(ctx context.Context, id, userID int, content string) (*domain.Comment, error)
	Delete(ctx context.Context, id, userID int, role string) error
	GetCount(ctx context.Context, videoID int) (int64, error)
}

type commentService struct {
	commentRepo repository.CommentRepository
	videoRepo   repository.VideoRepository
	countCache  repository.VideoStatsCache
	channelSvc  ChannelService
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	videoRepo repository.VideoRepository,
	countCache repository.VideoStatsCache,
	channelSvc ChannelService,
) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		videoRepo:   videoRepo,
		countCache:  countCache,
		channelSvc:  channelSvc,
	}
}

func (s *commentService) Create(ctx context.Context, userID, videoID int, content string) (*domain.Comment, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommentService"),
		slog.String("operation", "Create"),
		slog.Int("user_id", userID),
		slog.Int("video_id", videoID),
	)

	logger.DebugContext(ctx, "Checking video existence")
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get video: %w", err)
	}
	if video == nil {
		logger.WarnContext(ctx, "Video not found")
		return nil, domain.ErrVideoNotFound
	}

	if content == "" {
		logger.WarnContext(ctx, "Comment content is empty")
		return nil, fmt.Errorf("comment content cannot be empty")
	}

	comment := &domain.Comment{
		UserID:    userID,
		VideoID:   videoID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	logger.DebugContext(ctx, "Creating comment in repository")
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		logger.ErrorContext(ctx, "Failed to create comment", slog.String("error", err.Error()))
		return nil, fmt.Errorf("create comment: %w", err)
	}
	logger = logger.With(slog.Int("comment_id", comment.ID))

	_ = s.countCache.IncrComments(ctx, videoID)
	logger.InfoContext(ctx, "Comment created successfully")
	return comment, nil
}

func (s *commentService) GetByID(ctx context.Context, id int) (*domain.Comment, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommentService"),
		slog.String("operation", "GetByID"),
		slog.Int("comment_id", id),
	)

	logger.DebugContext(ctx, "Fetching comment by ID")
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get comment", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get comment: %w", err)
	}
	if comment == nil {
		logger.WarnContext(ctx, "Comment not found")
		return nil, domain.ErrCommentNotFound
	}
	logger.DebugContext(ctx, "Comment retrieved successfully")
	return comment, nil
}

func (s *commentService) ListByVideo(ctx context.Context, videoID int, limit, offset int) ([]*domain.Comment, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommentService"),
		slog.String("operation", "ListByVideo"),
		slog.Int("video_id", videoID),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	logger.DebugContext(ctx, "Checking video existence")
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get video: %w", err)
	}
	if video == nil {
		logger.WarnContext(ctx, "Video not found")
		return nil, domain.ErrVideoNotFound
	}

	logger.DebugContext(ctx, "Listing comments from repository")
	comments, err := s.commentRepo.ListByVideo(ctx, videoID, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list comments", slog.String("error", err.Error()))
		return nil, err
	}
	logger.DebugContext(ctx, "Comments listed successfully", slog.Int("count", len(comments)))
	return comments, nil
}

func (s *commentService) Update(ctx context.Context, id, userID int, content string) (*domain.Comment, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommentService"),
		slog.String("operation", "Update"),
		slog.Int("comment_id", id),
		slog.Int("user_id", userID),
	)

	logger.DebugContext(ctx, "Fetching comment for update")
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get comment", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get comment: %w", err)
	}
	if comment == nil {
		logger.WarnContext(ctx, "Comment not found")
		return nil, domain.ErrCommentNotFound
	}

	if comment.UserID != userID {
		logger.WarnContext(ctx, "User not allowed to edit comment", slog.Int("comment_owner_id", comment.UserID))
		return nil, domain.ErrForbidden
	}

	if content == "" {
		logger.WarnContext(ctx, "Comment content is empty")
		return nil, fmt.Errorf("comment content cannot be empty")
	}

	comment.Content = content
	logger.DebugContext(ctx, "Updating comment in repository")
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		logger.ErrorContext(ctx, "Failed to update comment", slog.String("error", err.Error()))
		return nil, fmt.Errorf("update comment: %w", err)
	}
	logger.InfoContext(ctx, "Comment updated successfully")
	return comment, nil
}

func (s *commentService) Delete(ctx context.Context, id, userID int, role string) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommentService"),
		slog.String("operation", "Delete"),
		slog.Int("comment_id", id),
		slog.Int("user_id", userID),
		slog.String("role", role),
	)

	logger.DebugContext(ctx, "Fetching comment for deletion")
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get comment", slog.String("error", err.Error()))
		return fmt.Errorf("get comment: %w", err)
	}
	if comment == nil {
		logger.WarnContext(ctx, "Comment not found")
		return domain.ErrCommentNotFound
	}
	logger = logger.With(slog.Int("comment_owner_id", comment.UserID), slog.Int("video_id", comment.VideoID))

	allowed := false
	if comment.UserID == userID {
		allowed = true
		logger.DebugContext(ctx, "User is comment author, allowed to delete")
	} else if role == "moderator" {
		allowed = true
		logger.DebugContext(ctx, "User is moderator, allowed to delete")
	}
	if !allowed {
		logger.DebugContext(ctx, "Checking if user is channel owner")
		video, err := s.videoRepo.GetByID(ctx, comment.VideoID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to get video for owner check", slog.String("error", err.Error()))
			return fmt.Errorf("get video: %w", err)
		}
		isOwner, err := s.channelSvc.IsOwner(ctx, video.ChannelID, userID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to check channel ownership", slog.String("error", err.Error()))
			return fmt.Errorf("check owner: %w", err)
		}
		if isOwner {
			allowed = true
			logger.DebugContext(ctx, "User is channel owner, allowed to delete")
		}
	}
	if !allowed {
		logger.WarnContext(ctx, "User not authorized to delete comment")
		return domain.ErrForbidden
	}

	logger.DebugContext(ctx, "Deleting comment from repository")
	if err := s.commentRepo.Delete(ctx, id); err != nil {
		logger.ErrorContext(ctx, "Failed to delete comment", slog.String("error", err.Error()))
		return fmt.Errorf("delete comment: %w", err)
	}

	_ = s.countCache.DecrComments(ctx, comment.VideoID)
	logger.InfoContext(ctx, "Comment deleted successfully")
	return nil
}

func (s *commentService) GetCount(ctx context.Context, videoID int) (int64, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommentService"),
		slog.String("operation", "GetCount"),
		slog.Int("video_id", videoID),
	)

	logger.DebugContext(ctx, "Trying to get comment count from cache")
	count, hit, err := s.countCache.GetCommentsCount(ctx, videoID)
	if err == nil && hit {
		logger.DebugContext(ctx, "Comment count retrieved from cache", slog.Int64("count", count))
		return count, nil
	}
	if err != nil {
		logger.WarnContext(ctx, "Cache error, falling back to DB", slog.String("error", err.Error()))
	}

	logger.DebugContext(ctx, "Fetching comment count from database")
	count, err = s.commentRepo.CountByVideo(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to count comments", slog.String("error", err.Error()))
		return 0, fmt.Errorf("count comments: %w", err)
	}

	_ = s.countCache.SetCommentsCount(ctx, videoID, count)
	logger.DebugContext(ctx, "Comment count retrieved from DB and cached", slog.Int64("count", count))
	return count, nil
}
