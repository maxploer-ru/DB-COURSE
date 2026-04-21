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

type PlaylistService interface {
	Create(ctx context.Context, channelID, userID int, name, description string) (*domain.Playlist, error)
	GetByID(ctx context.Context, playlistID int) (*domain.Playlist, error)
	ListByChannel(ctx context.Context, channelID int, limit, offset int) ([]*domain.Playlist, error)
	Update(ctx context.Context, playlistID, userID int, name, description *string) (*domain.Playlist, error)
	Delete(ctx context.Context, playlistID, userID int) error
	AddVideo(ctx context.Context, playlistID, videoID, userID int) error
	RemoveVideo(ctx context.Context, playlistID, videoID, userID int) error
}

type playlistService struct {
	playlistRepo repository.PlaylistRepository
	videoRepo    repository.VideoRepository
	channelSvc   ChannelService
}

func NewPlaylistService(
	playlistRepo repository.PlaylistRepository,
	videoRepo repository.VideoRepository,
	channelSvc ChannelService,
) PlaylistService {
	return &playlistService{
		playlistRepo: playlistRepo,
		videoRepo:    videoRepo,
		channelSvc:   channelSvc,
	}
}

func (s *playlistService) Create(ctx context.Context, channelID, userID int, name, description string) (*domain.Playlist, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "PlaylistService"),
		slog.String("operation", "Create"),
		slog.Int("channel_id", channelID),
		slog.Int("user_id", userID),
	)

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, domain.ErrPlaylistNameEmpty
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("check channel owner failed: %w", err)
	}
	if !isOwner {
		return nil, domain.ErrForbidden
	}

	playlist := &domain.Playlist{
		ChannelID:   channelID,
		Name:        name,
		Description: strings.TrimSpace(description),
		CreatedAt:   time.Now(),
	}
	if err := s.playlistRepo.Create(ctx, playlist); err != nil {
		logger.ErrorContext(ctx, "Failed to create playlist", slog.String("error", err.Error()))
		return nil, err
	}

	logger.InfoContext(ctx, "Playlist created", slog.Int("playlist_id", playlist.ID))
	return playlist, nil
}

func (s *playlistService) GetByID(ctx context.Context, playlistID int) (*domain.Playlist, error) {
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, err
	}
	if playlist == nil {
		return nil, domain.ErrPlaylistNotFound
	}
	return playlist, nil
}

func (s *playlistService) ListByChannel(ctx context.Context, channelID int, limit, offset int) ([]*domain.Playlist, error) {
	exists, err := s.channelSvc.Exists(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("check channel exists failed: %w", err)
	}
	if !exists {
		return nil, domain.ErrChannelNotFound
	}
	return s.playlistRepo.ListByChannel(ctx, channelID, limit, offset)
}

func (s *playlistService) Update(ctx context.Context, playlistID, userID int, name, description *string) (*domain.Playlist, error) {
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, err
	}
	if playlist == nil {
		return nil, domain.ErrPlaylistNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, playlist.ChannelID, userID)
	if err != nil {
		return nil, fmt.Errorf("check channel owner failed: %w", err)
	}
	if !isOwner {
		return nil, domain.ErrForbidden
	}

	updated := false
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return nil, domain.ErrPlaylistNameEmpty
		}
		if trimmed != playlist.Name {
			playlist.Name = trimmed
			updated = true
		}
	}
	if description != nil {
		playlist.Description = strings.TrimSpace(*description)
		updated = true
	}
	if !updated {
		return playlist, nil
	}

	if err := s.playlistRepo.Update(ctx, playlist); err != nil {
		return nil, err
	}
	return playlist, nil
}

func (s *playlistService) Delete(ctx context.Context, playlistID, userID int) error {
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return err
	}
	if playlist == nil {
		return domain.ErrPlaylistNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, playlist.ChannelID, userID)
	if err != nil {
		return fmt.Errorf("check channel owner failed: %w", err)
	}
	if !isOwner {
		return domain.ErrForbidden
	}

	return s.playlistRepo.Delete(ctx, playlistID)
}

func (s *playlistService) AddVideo(ctx context.Context, playlistID, videoID, userID int) error {
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return err
	}
	if playlist == nil {
		return domain.ErrPlaylistNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, playlist.ChannelID, userID)
	if err != nil {
		return fmt.Errorf("check channel owner failed: %w", err)
	}
	if !isOwner {
		return domain.ErrForbidden
	}

	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return fmt.Errorf("get video failed: %w", err)
	}
	if video == nil {
		return domain.ErrVideoNotFound
	}
	if video.ChannelID != playlist.ChannelID {
		return domain.ErrPlaylistVideoChannelMismatch
	}

	return s.playlistRepo.AddVideo(ctx, playlistID, videoID)
}

func (s *playlistService) RemoveVideo(ctx context.Context, playlistID, videoID, userID int) error {
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return err
	}
	if playlist == nil {
		return domain.ErrPlaylistNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, playlist.ChannelID, userID)
	if err != nil {
		return fmt.Errorf("check channel owner failed: %w", err)
	}
	if !isOwner {
		return domain.ErrForbidden
	}

	return s.playlistRepo.RemoveVideo(ctx, playlistID, videoID)
}
