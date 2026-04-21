package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type VideoService interface {
	CreateVideo(ctx context.Context, channelID, userID int, title, description, fileKey string) (*domain.Video, error)
	GetVideo(ctx context.Context, videoID int) (*domain.Video, error)
	UpdateVideo(ctx context.Context, videoID, userID int, title, description *string) (*domain.Video, error)
	DeleteVideo(ctx context.Context, videoID, userID int) error
	ListChannelVideos(ctx context.Context, channelID int, limit, offset int) ([]*domain.Video, error)
	ListMyVideos(ctx context.Context, userID int, limit, offset int) ([]*domain.Video, error)
	ListAllVideos(ctx context.Context, limit, offset int) ([]*domain.Video, error)
	GetUploadPresignedURL(ctx context.Context, channelID, userID int, filename string) (url string, fileKey string, err error)
	GetStreamingPresignedURL(ctx context.Context, videoID int) (string, error)
}

type StorageService interface {
	GenerateUploadPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	GenerateAccessPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	DeleteObject(ctx context.Context, key string) error
}

type videoService struct {
	videoRepo  repository.VideoRepository
	subRepo    repository.SubscriptionRepository
	channelSvc ChannelService
	storageSvc StorageService
}

func NewVideoService(
	videoRepo repository.VideoRepository,
	subRepo repository.SubscriptionRepository,
	channelSvc ChannelService,
	storageSvc StorageService,
) VideoService {
	return &videoService{
		videoRepo:  videoRepo,
		subRepo:    subRepo,
		channelSvc: channelSvc,
		storageSvc: storageSvc,
	}
}

func (s *videoService) CreateVideo(ctx context.Context, channelID, userID int, title, description, fileKey string) (*domain.Video, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoService"),
		slog.String("operation", "CreateVideo"),
		slog.Int("channel_id", channelID),
		slog.Int("user_id", userID),
		slog.String("title", title),
	)

	logger.DebugContext(ctx, "Checking channel ownership")
	isOwner, err := s.channelSvc.IsOwner(ctx, channelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel ownership", slog.String("error", err.Error()))
		return nil, fmt.Errorf("check channel owner: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not the channel owner")
		return nil, domain.ErrForbidden
	}

	video := &domain.Video{
		ChannelID:   channelID,
		Title:       title,
		Description: description,
		Filepath:    fileKey,
		CreatedAt:   time.Now(),
	}

	logger.DebugContext(ctx, "Creating video record in repository")
	if err := s.videoRepo.Create(ctx, video); err != nil {
		logger.ErrorContext(ctx, "Failed to create video in repository", slog.String("error", err.Error()))
		return nil, fmt.Errorf("create video failed: %w", err)
	}

	logger.InfoContext(ctx, "Video created successfully", slog.Int("video_id", video.ID))
	if err := s.subRepo.NotifySubscribersAboutNewVideo(ctx, channelID); err != nil {
		logger.WarnContext(ctx, "Failed to notify subscribers about new video", slog.String("error", err.Error()))
	}
	return video, nil
}

func (s *videoService) GetVideo(ctx context.Context, videoID int) (*domain.Video, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoService"),
		slog.String("operation", "GetVideo"),
		slog.Int("video_id", videoID),
	)

	logger.DebugContext(ctx, "Fetching video by ID")
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get video failed: %w", err)
	}
	if video == nil {
		logger.WarnContext(ctx, "Video not found")
		return nil, domain.ErrVideoNotFound
	}
	logger.DebugContext(ctx, "Video retrieved successfully")
	return video, nil
}

func (s *videoService) UpdateVideo(ctx context.Context, videoID, userID int, title, description *string) (*domain.Video, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoService"),
		slog.String("operation", "UpdateVideo"),
		slog.Int("video_id", videoID),
		slog.Int("user_id", userID),
	)

	logger.DebugContext(ctx, "Fetching video for update")
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get video failed: %w", err)
	}
	if video == nil {
		logger.WarnContext(ctx, "Video not found")
		return nil, domain.ErrVideoNotFound
	}

	logger.DebugContext(ctx, "Checking channel ownership")
	isOwner, err := s.channelSvc.IsOwner(ctx, video.ChannelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel ownership", slog.String("error", err.Error()))
		return nil, fmt.Errorf("check channel owner: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not the channel owner")
		return nil, domain.ErrForbidden
	}

	updated := false
	if title != nil {
		video.Title = *title
		updated = true
	}
	if description != nil {
		video.Description = *description
		updated = true
	}

	if !updated {
		logger.DebugContext(ctx, "No changes to update")
		return video, nil
	}

	logger.DebugContext(ctx, "Updating video in repository")
	if err := s.videoRepo.Update(ctx, video); err != nil {
		logger.ErrorContext(ctx, "Failed to update video", slog.String("error", err.Error()))
		return nil, fmt.Errorf("update video failed: %w", err)
	}

	logger.InfoContext(ctx, "Video updated successfully")
	return video, nil
}

func (s *videoService) DeleteVideo(ctx context.Context, videoID, userID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoService"),
		slog.String("operation", "DeleteVideo"),
		slog.Int("video_id", videoID),
		slog.Int("user_id", userID),
	)

	logger.DebugContext(ctx, "Fetching video for deletion")
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video", slog.String("error", err.Error()))
		return fmt.Errorf("get video failed: %w", err)
	}
	if video == nil {
		logger.WarnContext(ctx, "Video not found")
		return domain.ErrVideoNotFound
	}
	logger = logger.With(slog.Int("channel_id", video.ChannelID), slog.String("filepath", video.Filepath))

	logger.DebugContext(ctx, "Checking channel ownership")
	isOwner, err := s.channelSvc.IsOwner(ctx, video.ChannelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel ownership", slog.String("error", err.Error()))
		return fmt.Errorf("check channel owner: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not the channel owner")
		return domain.ErrForbidden
	}

	logger.DebugContext(ctx, "Deleting video file from storage")
	if err := s.storageSvc.DeleteObject(ctx, video.Filepath); err != nil {
		logger.WarnContext(ctx, "Failed to delete video file from storage", slog.String("error", err.Error()))
		// Продолжаем удаление записи из БД
	}

	logger.DebugContext(ctx, "Deleting video from repository")
	if err := s.videoRepo.Delete(ctx, videoID); err != nil {
		logger.ErrorContext(ctx, "Failed to delete video from repository", slog.String("error", err.Error()))
		return fmt.Errorf("delete video failed: %w", err)
	}

	logger.InfoContext(ctx, "Video deleted successfully")
	return nil
}

func (s *videoService) ListChannelVideos(ctx context.Context, channelID int, limit, offset int) ([]*domain.Video, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoService"),
		slog.String("operation", "ListChannelVideos"),
		slog.Int("channel_id", channelID),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	logger.DebugContext(ctx, "Checking channel existence")
	exists, err := s.channelSvc.Exists(ctx, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel existence", slog.String("error", err.Error()))
		return nil, fmt.Errorf("check channel exists: %w", err)
	}
	if !exists {
		logger.WarnContext(ctx, "Channel not found")
		return nil, domain.ErrChannelNotFound
	}

	logger.DebugContext(ctx, "Listing videos from repository")
	videos, err := s.videoRepo.ListByChannel(ctx, channelID, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list videos", slog.String("error", err.Error()))
		return nil, fmt.Errorf("list videos failed: %w", err)
	}
	logger.DebugContext(ctx, "Videos listed successfully", slog.Int("count", len(videos)))
	return videos, nil
}

func (s *videoService) ListMyVideos(ctx context.Context, userID int, limit, offset int) ([]*domain.Video, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoService"),
		slog.String("operation", "ListMyVideos"),
		slog.Int("user_id", userID),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	logger.DebugContext(ctx, "Fetching user's channel")
	channel, err := s.channelSvc.GetChannelByUserID(ctx, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get channel by user ID", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get channel by user id failed: %w", err)
	}
	logger = logger.With(slog.Int("channel_id", channel.ID))

	logger.DebugContext(ctx, "Listing user's videos")
	videos, err := s.videoRepo.ListByChannel(ctx, channel.ID, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list videos", slog.String("error", err.Error()))
		return nil, err
	}
	logger.DebugContext(ctx, "Videos listed successfully", slog.Int("count", len(videos)))
	return videos, nil
}

func (s *videoService) ListAllVideos(ctx context.Context, limit, offset int) ([]*domain.Video, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoService"),
		slog.String("operation", "ListAllVideos"),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	logger.DebugContext(ctx, "Listing all videos")
	videos, err := s.videoRepo.List(ctx, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list videos", slog.String("error", err.Error()))
		return nil, fmt.Errorf("list videos failed: %w", err)
	}
	logger.DebugContext(ctx, "Videos listed successfully", slog.Int("count", len(videos)))
	return videos, nil
}

func (s *videoService) GetUploadPresignedURL(ctx context.Context, channelID, userID int, filename string) (string, string, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoService"),
		slog.String("operation", "GetUploadPresignedURL"),
		slog.Int("channel_id", channelID),
		slog.Int("user_id", userID),
		slog.String("filename", filename),
	)

	logger.DebugContext(ctx, "Checking channel ownership")
	isOwner, err := s.channelSvc.IsOwner(ctx, channelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel ownership", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("check channel owner: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not the channel owner")
		return "", "", domain.ErrForbidden
	}

	fileKey := fmt.Sprintf("videos/%d/%d_%s", channelID, time.Now().UnixNano(), filename)
	logger = logger.With(slog.String("file_key", fileKey))

	logger.DebugContext(ctx, "Generating upload presigned URL")
	url, err := s.storageSvc.GenerateUploadPresignedURL(ctx, fileKey, 15*time.Minute)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to generate upload URL", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("generate upload url failed: %w", err)
	}

	logger.InfoContext(ctx, "Upload presigned URL generated successfully")
	return url, fileKey, nil
}

func (s *videoService) GetStreamingPresignedURL(ctx context.Context, videoID int) (string, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoService"),
		slog.String("operation", "GetStreamingPresignedURL"),
		slog.Int("video_id", videoID),
	)

	logger.DebugContext(ctx, "Fetching video")
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video", slog.String("error", err.Error()))
		return "", fmt.Errorf("get video failed: %w", err)
	}
	if video == nil {
		logger.WarnContext(ctx, "Video not found")
		return "", domain.ErrVideoNotFound
	}
	logger = logger.With(slog.String("file_key", video.Filepath))

	logger.DebugContext(ctx, "Generating streaming presigned URL")
	url, err := s.storageSvc.GenerateAccessPresignedURL(ctx, video.Filepath, 1*time.Hour)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to generate streaming URL", slog.String("error", err.Error()))
		return "", fmt.Errorf("generate streaming url failed: %w", err)
	}

	logger.InfoContext(ctx, "Streaming presigned URL generated successfully")
	return url, nil
}
