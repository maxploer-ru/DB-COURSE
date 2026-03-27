package service

import (
	"ZVideo/internal/domain/video/entity"
	"ZVideo/internal/domain/video/repository"
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrVideoNotFound   = errors.New("video not found")
	ErrInvalidTitle    = errors.New("title must be 1 to 100 characters long")
	ErrChannelNotFound = errors.New("channel not found")
	ErrNotAuthorized   = errors.New("not authorized")
	ErrRatingNotFound  = errors.New("rating not found")
)

type ChannelChecker interface {
	Exists(ctx context.Context, channelID int) (bool, error)
	IsOwner(ctx context.Context, channelID, userID int) (bool, error)
}

type VideoService interface {
	CreateVideo(ctx context.Context, channelID, userID int, title, description, filepath string) (*entity.Video, error)

	GetVideo(ctx context.Context, videoID int) (*entity.Video, error)

	UpdateVideo(ctx context.Context, videoID, userID int, title, description *string) (*entity.Video, error)

	DeleteVideo(ctx context.Context, videoID, userID int) error

	RecordViewing(ctx context.Context, userID, videoID int) error

	GetChannelVideos(ctx context.Context, channelID int, limit, offset int) ([]*entity.Video, error)

	SearchVideos(ctx context.Context, query string, limit, offset int) ([]*entity.Video, error)
}

type videoService struct {
	videoRepo    repository.VideoRepository
	viewingRepo  repository.ViewingRepository
	statsCache   repository.VideoStatsCache
	channelCheck ChannelChecker
}

func NewVideoService(
	videoRepo repository.VideoRepository,
	viewingRepo repository.ViewingRepository,
	statsCache repository.VideoStatsCache,
	channelCheck ChannelChecker,
) VideoService {
	return &videoService{
		videoRepo:    videoRepo,
		viewingRepo:  viewingRepo,
		statsCache:   statsCache,
		channelCheck: channelCheck,
	}
}

func (s *videoService) CreateVideo(ctx context.Context, channelID, userID int, title, description, filepath string) (*entity.Video, error) {

	if title == "" || len(title) > 100 {
		return nil, ErrInvalidTitle
	}

	exists, err := s.channelCheck.Exists(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("check channel exists: %w", err)
	}
	if !exists {
		return nil, ErrChannelNotFound
	}
	isOwner, err := s.channelCheck.IsOwner(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("check owner: %w", err)
	}
	if !isOwner {
		return nil, ErrNotAuthorized
	}

	video := &entity.Video{
		ChannelID:   channelID,
		Title:       title,
		Description: description,
		Filepath:    filepath,
	}
	if err := s.videoRepo.Create(ctx, video); err != nil {
		return nil, fmt.Errorf("create video: %w", err)
	}
	return video, nil
}

func (s *videoService) GetVideo(ctx context.Context, videoID int) (*entity.Video, error) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("get video: %w", err)
	}
	if video == nil {
		return nil, ErrVideoNotFound
	}
	return video, nil
}

func (s *videoService) UpdateVideo(ctx context.Context, videoID, userID int, title, description *string) (*entity.Video, error) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("get video: %w", err)
	}
	if video == nil {
		return nil, ErrVideoNotFound
	}

	isOwner, err := s.channelCheck.IsOwner(ctx, video.ChannelID, userID)
	if err != nil {
		return nil, fmt.Errorf("check owner: %w", err)
	}
	if !isOwner {
		return nil, ErrNotAuthorized
	}

	updated := false
	if title != nil && *title != video.Title {
		if *title == "" || len(*title) > 100 {
			return nil, ErrInvalidTitle
		}
		video.Title = *title
		updated = true
	}
	if description != nil && *description != video.Description {
		video.Description = *description
		updated = true
	}

	if updated {
		if err := s.videoRepo.Update(ctx, video); err != nil {
			return nil, fmt.Errorf("update video: %w", err)
		}
	}
	return video, nil
}

func (s *videoService) DeleteVideo(ctx context.Context, videoID, userID int) error {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return err
	}
	if video == nil {
		return ErrVideoNotFound
	}

	isOwner, err := s.channelCheck.IsOwner(ctx, video.ChannelID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return ErrNotAuthorized
	}

	if err := s.videoRepo.Delete(ctx, videoID); err != nil {
		return err
	}

	_ = s.statsCache.Invalidate(ctx, videoID)
	return nil
}

func (s *videoService) RecordViewing(ctx context.Context, userID, videoID int) error {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return err
	}
	if video == nil {
		return ErrVideoNotFound
	}

	viewing := &entity.Viewing{
		UserID:    userID,
		VideoID:   videoID,
		WatchedAt: time.Now(),
	}

	if err := s.viewingRepo.Create(ctx, viewing); err != nil {
		return err
	}

	_ = s.statsCache.Invalidate(ctx, videoID)
	return nil
}

func (s *videoService) GetChannelVideos(ctx context.Context, channelID int, limit, offset int) ([]*entity.Video, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.videoRepo.GetByChannelID(ctx, channelID, limit, offset)
}

func (s *videoService) SearchVideos(ctx context.Context, query string, limit, offset int) ([]*entity.Video, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.videoRepo.Search(ctx, query, limit, offset)
}
