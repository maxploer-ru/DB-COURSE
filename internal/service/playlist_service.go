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
	ListByChannel(ctx context.Context, channelID int, limit, offset int) (*domain.PageResponse[*domain.Playlist], error)
	Update(ctx context.Context, playlistID, userID int, name, description *string) (*domain.Playlist, error)
	Delete(ctx context.Context, playlistID, userID int) error

	AddVideo(ctx context.Context, playlistID, videoID, userID int) error
	RemoveVideo(ctx context.Context, playlistID, videoID, userID int) error
	UpdateVideoPosition(ctx context.Context, playlistID, videoID, userID, newPosition int) error

	GetMyPlaylists(ctx context.Context, userID int, limit, offset int) (*domain.PageResponse[*domain.Playlist], error)
	GetPlaylistItems(ctx context.Context, playlistID int, limit, offset int) (*domain.PageResponse[*domain.PlaylistItem], error)
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

func (s *playlistService) ListByChannel(ctx context.Context, channelID int, limit, offset int) (*domain.PageResponse[*domain.Playlist], error) {
	exists, err := s.channelSvc.Exists(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("check channel exists failed: %w", err)
	}
	if !exists {
		return nil, domain.ErrChannelNotFound
	}
	playlists, err := s.playlistRepo.ListByChannel(ctx, channelID, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.playlistRepo.CountByChannel(ctx, channelID)
	if err != nil {
		return nil, err
	}
	return &domain.PageResponse[*domain.Playlist]{
		Items:      playlists,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}, nil
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

func (s *playlistService) GetMyPlaylists(ctx context.Context, userID int, limit, offset int) (*domain.PageResponse[*domain.Playlist], error) {
	logger := domain.GetLogger(ctx).With(
		"service", "PlaylistService",
		"operation", "GetMyPlaylists",
		"user_id", userID,
	)

	logger.DebugContext(ctx, "Fetching user's channel")
	channel, err := s.channelSvc.GetChannelByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get channel by user id failed: %w", err)
	}
	if channel == nil {
		return &domain.PageResponse[*domain.Playlist]{
			Items:      []*domain.Playlist{},
			TotalCount: 0,
			Limit:      limit,
			Offset:     offset,
		}, nil
	}

	logger.DebugContext(ctx, "Listing playlists from repository")
	return s.ListByChannel(ctx, channel.ID, limit, offset)
}

func (s *playlistService) GetPlaylistItems(ctx context.Context, playlistID int, limit, offset int) (*domain.PageResponse[*domain.PlaylistItem], error) {
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, err
	}
	if playlist == nil {
		return nil, domain.ErrPlaylistNotFound
	}

	items, err := s.playlistRepo.ListItems(ctx, playlistID, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.playlistRepo.GetItemsCount(ctx, playlistID)
	if err != nil {
		return nil, err
	}
	return &domain.PageResponse[*domain.PlaylistItem]{
		Items:      items,
		TotalCount: int64(total),
		Limit:      limit,
		Offset:     offset,
	}, nil
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
	if err != nil || !isOwner {
		return domain.ErrForbidden
	}

	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil || video == nil {
		return domain.ErrVideoNotFound
	}

	if video.Status != domain.VideoStatusReady {
		return fmt.Errorf("cannot add pending video to playlist")
	}

	return s.playlistRepo.AddVideo(ctx, playlistID, videoID)
}

func (s *playlistService) RemoveVideo(ctx context.Context, playlistID, videoID, userID int) error {
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil || playlist == nil {
		return domain.ErrPlaylistNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, playlist.ChannelID, userID)
	if err != nil || !isOwner {
		return domain.ErrForbidden
	}

	return s.playlistRepo.RemoveVideo(ctx, playlistID, videoID)
}

func (s *playlistService) UpdateVideoPosition(ctx context.Context, playlistID, videoID, userID, newPosition int) error {
	logger := domain.GetLogger(ctx).With(
		"service", "PlaylistService",
		"operation", "UpdateVideoPosition",
		"playlist_id", playlistID,
		"video_id", videoID,
		"new_position", newPosition,
	)

	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil || playlist == nil {
		return domain.ErrPlaylistNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, playlist.ChannelID, userID)
	if err != nil || !isOwner {
		return domain.ErrForbidden
	}

	maxPos, err := s.playlistRepo.GetItemsCount(ctx, playlistID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get playlist items count", slog.String("error", err.Error()))
		return err
	}

	if newPosition < 1 {
		newPosition = 1
	}
	if newPosition > maxPos {
		newPosition = maxPos
	}

	logger.DebugContext(ctx, "Updating video position in repository")
	return s.playlistRepo.UpdateVideoPosition(ctx, playlistID, videoID, newPosition)
}
