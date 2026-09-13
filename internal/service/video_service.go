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
	ListChannelVideos(ctx context.Context, channelID int, limit, offset int) ([]*domain.Video, error)
	ListMyVideos(ctx context.Context, userID int, limit, offset int) ([]*domain.Video, error)
	ListAllVideos(ctx context.Context, limit, offset int) ([]*domain.Video, error)
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
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoService"),
		slog.String("operation", "InitUpload"),
		slog.Int("channel_id", channelID),
	)

	isOwner, err := s.channelSvc.IsOwner(ctx, channelID, userID)
	if err != nil {
		return nil, "", fmt.Errorf("check channel owner: %w", err)
	}
	if !isOwner {
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
		return nil, "", fmt.Errorf("generate upload url failed: %w", err)
	}

	return video, url, nil
}

func (s *videoService) ConfirmUpload(ctx context.Context, videoID, userID int) error {
	logger := domain.GetLogger(ctx).With("operation", "ConfirmUpload", "video_id", videoID)

	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return fmt.Errorf("get video: %w", err)
	}
	if video == nil {
		return domain.ErrVideoNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, video.ChannelID, userID)
	if err != nil || !isOwner {
		return domain.ErrForbidden
	}

	if video.Status == domain.VideoStatusReady {
		return nil
	}

	video.Status = domain.VideoStatusReady
	if err := s.videoRepo.Update(ctx, video); err != nil {
		return fmt.Errorf("update video status: %w", err)
	}

	if err := s.subSvc.NotifyAboutNewVideo(ctx, video.ChannelID); err != nil {
		logger.WarnContext(ctx, "Failed to notify subscribers", slog.String("error", err.Error()))
	}

	logger.InfoContext(ctx, "Video confirmed and published")
	return nil
}

func (s *videoService) DeleteVideo(ctx context.Context, videoID, userID int, role string) error {
	logger := domain.GetLogger(ctx).With("operation", "DeleteVideo", "video_id", videoID)

	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return fmt.Errorf("get video failed: %w", err)
	}
	if video == nil {
		return domain.ErrVideoNotFound
	}

	allowed := role == domain.RoleModerator || role == domain.RoleAdmin
	if !allowed {
		isOwner, err := s.channelSvc.IsOwner(ctx, video.ChannelID, userID)
		if err != nil {
			return fmt.Errorf("check channel owner: %w", err)
		}
		if !isOwner {
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
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("get video failed: %w", err)
	}
	if video == nil {
		return nil, domain.ErrVideoNotFound
	}
	return video, nil
}

func (s *videoService) UpdateVideo(ctx context.Context, videoID, userID int, title, description *string) (*domain.Video, error) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil || video == nil {
		return nil, domain.ErrVideoNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, video.ChannelID, userID)
	if err != nil || !isOwner {
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
			return nil, fmt.Errorf("update video failed: %w", err)
		}
	}
	return video, nil
}

func (s *videoService) ListChannelVideos(ctx context.Context, channelID int, limit, offset int) ([]*domain.Video, error) {
	exists, err := s.channelSvc.Exists(ctx, channelID)
	if err != nil || !exists {
		return nil, domain.ErrChannelNotFound
	}
	return s.videoRepo.ListByChannel(ctx, channelID, limit, offset)
}

func (s *videoService) ListMyVideos(ctx context.Context, userID int, limit, offset int) ([]*domain.Video, error) {
	channel, err := s.channelSvc.GetChannelByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get channel by user id: %w", err)
	}
	return s.videoRepo.ListByChannel(ctx, channel.ID, limit, offset)
}

func (s *videoService) ListAllVideos(ctx context.Context, limit, offset int) ([]*domain.Video, error) {
	return s.videoRepo.List(ctx, limit, offset)
}

func (s *videoService) GetStreamingPresignedURL(ctx context.Context, videoID int) (string, error) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil || video == nil {
		return "", domain.ErrVideoNotFound
	}

	if video.Status != domain.VideoStatusReady {
		return "", fmt.Errorf("video is not ready yet")
	}

	url, err := s.storageSvc.GenerateAccessPresignedURL(ctx, video.Filepath, 1*time.Hour)
	if err != nil {
		return "", fmt.Errorf("generate streaming url failed: %w", err)
	}

	return url, nil
}
