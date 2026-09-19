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
	logger := serviceLogger(ctx, "PlaylistService", "GetByID", slog.Int("playlist_id", playlistID))
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get playlist", slog.Any("error", err))
		return nil, err
	}
	if playlist == nil {
		logger.WarnContext(ctx, "Playlist not found")
		return nil, domain.ErrPlaylistNotFound
	}
	return playlist, nil
}

func (s *playlistService) ListByChannel(ctx context.Context, channelID int, limit, offset int) (*domain.PageResponse[*domain.Playlist], error) {
	logger := serviceLogger(ctx, "PlaylistService", "ListByChannel",
		slog.Int("channel_id", channelID), slog.Int("limit", limit), slog.Int("offset", offset))
	exists, err := s.channelSvc.Exists(ctx, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel existence", slog.Any("error", err))
		return nil, fmt.Errorf("check channel exists failed: %w", err)
	}
	if !exists {
		logger.WarnContext(ctx, "Channel not found")
		return nil, domain.ErrChannelNotFound
	}
	playlists, err := s.playlistRepo.ListByChannel(ctx, channelID, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list playlists", slog.Any("error", err))
		return nil, err
	}
	total, err := s.playlistRepo.CountByChannel(ctx, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to count playlists", slog.Any("error", err))
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
	logger := serviceLogger(ctx, "PlaylistService", "Update",
		slog.Int("playlist_id", playlistID), slog.Int("user_id", userID))
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get playlist for update", slog.Any("error", err))
		return nil, err
	}
	if playlist == nil {
		logger.WarnContext(ctx, "Playlist not found for update")
		return nil, domain.ErrPlaylistNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, playlist.ChannelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check playlist ownership", slog.Any("error", err))
		return nil, fmt.Errorf("check channel owner failed: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not the playlist owner")
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
		logger.DebugContext(ctx, "No playlist changes to update")
		return playlist, nil
	}

	if err := s.playlistRepo.Update(ctx, playlist); err != nil {
		logger.ErrorContext(ctx, "Failed to update playlist", slog.Any("error", err))
		return nil, err
	}
	logger.InfoContext(ctx, "Playlist updated")
	return playlist, nil
}

func (s *playlistService) Delete(ctx context.Context, playlistID, userID int) error {
	logger := serviceLogger(ctx, "PlaylistService", "Delete",
		slog.Int("playlist_id", playlistID), slog.Int("user_id", userID))
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get playlist for deletion", slog.Any("error", err))
		return err
	}
	if playlist == nil {
		logger.WarnContext(ctx, "Playlist not found for deletion")
		return domain.ErrPlaylistNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, playlist.ChannelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check playlist ownership", slog.Any("error", err))
		return fmt.Errorf("check channel owner failed: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not the playlist owner")
		return domain.ErrForbidden
	}

	if err := s.playlistRepo.Delete(ctx, playlistID); err != nil {
		logger.ErrorContext(ctx, "Failed to delete playlist", slog.Any("error", err))
		return err
	}
	logger.InfoContext(ctx, "Playlist deleted")
	return nil
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
		logger.ErrorContext(ctx, "Failed to get user's channel", slog.Any("error", err))
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
	logger := serviceLogger(ctx, "PlaylistService", "GetPlaylistItems",
		slog.Int("playlist_id", playlistID), slog.Int("limit", limit), slog.Int("offset", offset))
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get playlist", slog.Any("error", err))
		return nil, err
	}
	if playlist == nil {
		logger.WarnContext(ctx, "Playlist not found")
		return nil, domain.ErrPlaylistNotFound
	}

	items, err := s.playlistRepo.ListItems(ctx, playlistID, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list playlist items", slog.Any("error", err))
		return nil, err
	}
	total, err := s.playlistRepo.GetItemsCount(ctx, playlistID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to count playlist items", slog.Any("error", err))
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
	logger := serviceLogger(ctx, "PlaylistService", "AddVideo",
		slog.Int("playlist_id", playlistID), slog.Int("video_id", videoID), slog.Int("user_id", userID))
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get playlist", slog.Any("error", err))
		return err
	}
	if playlist == nil {
		logger.WarnContext(ctx, "Playlist not found")
		return domain.ErrPlaylistNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, playlist.ChannelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check playlist ownership", slog.Any("error", err))
		return fmt.Errorf("check channel owner failed: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not the playlist owner")
		return domain.ErrForbidden
	}

	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video", slog.Any("error", err))
		return fmt.Errorf("get video failed: %w", err)
	}
	if video == nil {
		logger.WarnContext(ctx, "Video not found")
		return domain.ErrVideoNotFound
	}

	if video.Status != domain.VideoStatusReady {
		logger.WarnContext(ctx, "Video is not ready")
		return domain.ErrVideoNotReady
	}
	if video.ChannelID != playlist.ChannelID {
		logger.WarnContext(ctx, "Video belongs to another channel")
		return domain.ErrPlaylistVideoChannelMismatch
	}

	if err := s.playlistRepo.AddVideo(ctx, playlistID, videoID); err != nil {
		logger.ErrorContext(ctx, "Failed to add video to playlist", slog.Any("error", err))
		return err
	}
	logger.InfoContext(ctx, "Video added to playlist")
	return nil
}

func (s *playlistService) RemoveVideo(ctx context.Context, playlistID, videoID, userID int) error {
	logger := serviceLogger(ctx, "PlaylistService", "RemoveVideo",
		slog.Int("playlist_id", playlistID), slog.Int("video_id", videoID), slog.Int("user_id", userID))
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get playlist", slog.Any("error", err))
		return fmt.Errorf("get playlist failed: %w", err)
	}
	if playlist == nil {
		logger.WarnContext(ctx, "Playlist not found")
		return domain.ErrPlaylistNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, playlist.ChannelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check playlist ownership", slog.Any("error", err))
		return fmt.Errorf("check channel owner failed: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not the playlist owner")
		return domain.ErrForbidden
	}

	if err := s.playlistRepo.RemoveVideo(ctx, playlistID, videoID); err != nil {
		logger.ErrorContext(ctx, "Failed to remove video from playlist", slog.Any("error", err))
		return err
	}
	logger.InfoContext(ctx, "Video removed from playlist")
	return nil
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
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get playlist for position update", slog.Any("error", err))
		return fmt.Errorf("get playlist failed: %w", err)
	}
	if playlist == nil {
		logger.WarnContext(ctx, "Playlist not found for position update")
		return domain.ErrPlaylistNotFound
	}

	isOwner, err := s.channelSvc.IsOwner(ctx, playlist.ChannelID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check playlist ownership", slog.Any("error", err))
		return fmt.Errorf("check channel owner failed: %w", err)
	}
	if !isOwner {
		logger.WarnContext(ctx, "User is not the playlist owner")
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
	if err := s.playlistRepo.UpdateVideoPosition(ctx, playlistID, videoID, newPosition); err != nil {
		logger.ErrorContext(ctx, "Failed to update video position", slog.Any("error", err))
		return err
	}
	logger.InfoContext(ctx, "Playlist video position updated")
	return nil
}
