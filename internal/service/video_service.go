package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type VideoService interface {
	InitUpload(ctx context.Context, channelID, userID int, title, description, filename string) (*domain.Video, string, error)
	ConfirmUpload(ctx context.Context, videoID, userID int) error

	GetVideo(ctx context.Context, videoID int) (*domain.Video, error)
	UpdateVideo(ctx context.Context, videoID, userID int, title, description *string) (*domain.Video, error)
	DeleteVideo(ctx context.Context, videoID, userID int, role string) error
	ListChannelVideos(ctx context.Context, channelID int, limit, offset int) (*domain.PageResponse[*domain.Video], error)
	ListMyVideos(ctx context.Context, userID int, limit, offset int) (*domain.PageResponse[*domain.Video], error)
	ListAllVideos(ctx context.Context, limit, offset int) (*domain.PageResponse[*domain.Video], error)
	GetStreamingPresignedURL(ctx context.Context, videoID int) (string, error)
}

type StorageService interface {
	GenerateUploadPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	GenerateAccessPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

type videoService struct {
	videoRepo  repository.VideoRepository
	subSvc     SubscriptionService
	channelSvc ChannelService
	storageSvc StorageService
}

func NewVideoService(
	videoRepo repository.VideoRepository,
	subSvc SubscriptionService,
	channelSvc ChannelService,
	storageSvc StorageService,
) VideoService {
	return &videoService{
		videoRepo:  videoRepo,
		subSvc:     subSvc,
		channelSvc: channelSvc,
		storageSvc: storageSvc,
	}
}

func (s *videoService) InitUpload(ctx context.Context, channelID, userID int, title, description, filename string) (*domain.Video, string, error) {
	logger := serviceLogger(ctx, "VideoService", "InitUpload",
		slog.Int("channel_id", channelID),
		slog.Int("user_id", userID),
	)

	isOwner, err := s.channelSvc.IsOwner(ctx, channelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel ownership", slog.Any("error", err))
		return nil, "", fmt.Errorf("check channel owner: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not the channel owner")
		return nil, "", domain.ErrForbidden
	}

	ext := strings.ToLower(filepath.Ext(filename))

	fileKey := fmt.Sprintf("videos/%d/%s%s", channelID, uuid.New().String(), ext)

	video := &domain.Video{
		ChannelID:        channelID,
		Title:            title,
		Description:      description,
		OriginalFilename: filename,
		Filepath:         fileKey,
		Status:           domain.VideoStatusPending,
		CreatedAt:        time.Now(),
	}

	if err := s.videoRepo.Create(ctx, video); err != nil {
		logger.ErrorContext(ctx, "Failed to create pending video", slog.String("error", err.Error()))
		return nil, "", fmt.Errorf("create video failed: %w", err)
	}

	url, err := s.storageSvc.GenerateUploadPresignedURL(ctx, fileKey, 15*time.Minute)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to generate upload URL", slog.Any("error", err))
		return nil, "", fmt.Errorf("generate upload url failed: %w", err)
	}

	logger.InfoContext(ctx, "Video upload initialized", slog.Int("video_id", video.ID))
	return video, url, nil
}

func (s *videoService) ConfirmUpload(ctx context.Context, videoID, userID int) error {
	logger := serviceLogger(ctx, "VideoService", "ConfirmUpload",
		slog.Int("video_id", videoID), slog.Int("user_id", userID))

	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video for confirmation", slog.Any("error", err))
		return fmt.Errorf("get video: %w", err)
	}
	if video == nil {
		logger.WarnContext(ctx, "Video not found for confirmation")
		return domain.ErrVideoNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, video.ChannelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel ownership", slog.Any("error", err))
		return fmt.Errorf("check channel owner: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not the channel owner")
		return domain.ErrForbidden
	}

	if video.Status == domain.VideoStatusReady {
		return nil
	}

	video.Status = domain.VideoStatusReady
	if err := s.videoRepo.Update(ctx, video); err != nil {
		logger.ErrorContext(ctx, "Failed to publish video", slog.Any("error", err))
		return fmt.Errorf("update video status: %w", err)
	}

	if err := s.subSvc.NotifyAboutNewVideo(ctx, video.ChannelID); err != nil {
		logger.WarnContext(ctx, "Failed to notify subscribers", slog.String("error", err.Error()))
	}

	logger.InfoContext(ctx, "Video confirmed and published")
	return nil
}

func (s *videoService) DeleteVideo(ctx context.Context, videoID, userID int, role string) error {
	logger := serviceLogger(ctx, "VideoService", "DeleteVideo",
		slog.Int("video_id", videoID), slog.Int("user_id", userID), slog.String("role", role))

	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video for deletion", slog.Any("error", err))
		return fmt.Errorf("get video failed: %w", err)
	}
	if video == nil {
		logger.WarnContext(ctx, "Video not found for deletion")
		return domain.ErrVideoNotFound
	}

	allowed := strings.EqualFold(role, domain.RoleModerator) || strings.EqualFold(role, domain.RoleAdmin)
	if !allowed {
		isOwner, err := s.channelSvc.IsOwner(ctx, video.ChannelID, userID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to check channel ownership", slog.Any("error", err))
			return fmt.Errorf("check channel owner: %w", err)
		}
		if !isOwner {
			logger.WarnContext(ctx, "User is not allowed to delete video")
			return domain.ErrForbidden
		}
	}

	if err := s.videoRepo.Delete(ctx, videoID); err != nil {
		logger.ErrorContext(ctx, "Failed to delete video from DB", slog.String("error", err.Error()))
		return fmt.Errorf("delete video failed: %w", err)
	}

	logger.InfoContext(ctx, "Video deleted successfully, S3 deletion queued by DB trigger")
	return nil
}

func (s *videoService) GetVideo(ctx context.Context, videoID int) (*domain.Video, error) {
	logger := serviceLogger(ctx, "VideoService", "GetVideo", slog.Int("video_id", videoID))
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video", slog.Any("error", err))
		return nil, fmt.Errorf("get video failed: %w", err)
	}
	if video == nil {
		logger.WarnContext(ctx, "Video not found")
		return nil, domain.ErrVideoNotFound
	}
	return video, nil
}

func (s *videoService) UpdateVideo(ctx context.Context, videoID, userID int, title, description *string) (*domain.Video, error) {
	logger := serviceLogger(ctx, "VideoService", "UpdateVideo",
		slog.Int("video_id", videoID), slog.Int("user_id", userID))
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video for update", slog.Any("error", err))
		return nil, fmt.Errorf("get video: %w", err)
	}
	if video == nil {
		logger.WarnContext(ctx, "Video not found for update")
		return nil, domain.ErrVideoNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, video.ChannelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel ownership", slog.Any("error", err))
		return nil, fmt.Errorf("check channel owner: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not allowed to update video")
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

	if updated {
		if err := s.videoRepo.Update(ctx, video); err != nil {
			logger.ErrorContext(ctx, "Failed to update video", slog.Any("error", err))
			return nil, fmt.Errorf("update video failed: %w", err)
		}
		logger.InfoContext(ctx, "Video updated")
	} else {
		logger.DebugContext(ctx, "No video changes to update")
	}
	return video, nil
}

func (s *videoService) ListChannelVideos(ctx context.Context, channelID int, limit, offset int) (*domain.PageResponse[*domain.Video], error) {
	logger := serviceLogger(ctx, "VideoService", "ListChannelVideos",
		slog.Int("channel_id", channelID), slog.Int("limit", limit), slog.Int("offset", offset))
	exists, err := s.channelSvc.Exists(ctx, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel existence", slog.Any("error", err))
		return nil, fmt.Errorf("check channel exists: %w", err)
	}
	if !exists {
		logger.WarnContext(ctx, "Channel not found")
		return nil, domain.ErrChannelNotFound
	}
	videos, err := s.videoRepo.ListByChannel(ctx, channelID, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list channel videos", slog.Any("error", err))
		return nil, err
	}
	total, err := s.videoRepo.CountByChannel(ctx, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to count channel videos", slog.Any("error", err))
		return nil, err
	}
	return &domain.PageResponse[*domain.Video]{
		Items:      videos,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func (s *videoService) ListMyVideos(ctx context.Context, userID int, limit, offset int) (*domain.PageResponse[*domain.Video], error) {
	logger := serviceLogger(ctx, "VideoService", "ListMyVideos",
		slog.Int("user_id", userID), slog.Int("limit", limit), slog.Int("offset", offset))
	channel, err := s.channelSvc.GetChannelByUserID(ctx, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get user's channel", slog.Any("error", err))
		return nil, fmt.Errorf("get channel by user id: %w", err)
	}
	if channel == nil {
		logger.WarnContext(ctx, "User has no channel")
		return nil, domain.ErrChannelNotFound
	}
	return s.ListChannelVideos(ctx, channel.ID, limit, offset)
}

func (s *videoService) ListAllVideos(ctx context.Context, limit, offset int) (*domain.PageResponse[*domain.Video], error) {
	logger := serviceLogger(ctx, "VideoService", "ListAllVideos",
		slog.Int("limit", limit), slog.Int("offset", offset))
	videos, err := s.videoRepo.List(ctx, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list videos", slog.Any("error", err))
		return nil, err
	}
	total, err := s.videoRepo.Count(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to count videos", slog.Any("error", err))
		return nil, err
	}
	return &domain.PageResponse[*domain.Video]{
		Items:      videos,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func (s *videoService) GetStreamingPresignedURL(ctx context.Context, videoID int) (string, error) {
	logger := serviceLogger(ctx, "VideoService", "GetStreamingPresignedURL", slog.Int("video_id", videoID))
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video for streaming", slog.Any("error", err))
		return "", fmt.Errorf("get video: %w", err)
	}
	if video == nil {
		logger.WarnContext(ctx, "Video not found for streaming")
		return "", domain.ErrVideoNotFound
	}

	if video.Status != domain.VideoStatusReady {
		logger.WarnContext(ctx, "Video is not ready for streaming")
		return "", domain.ErrVideoNotReady
	}

	url, err := s.storageSvc.GenerateAccessPresignedURL(ctx, video.Filepath, 1*time.Hour)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to generate streaming URL", slog.Any("error", err))
		return "", fmt.Errorf("generate streaming url failed: %w", err)
	}

	logger.DebugContext(ctx, "Streaming URL generated")
	return url, nil
}
