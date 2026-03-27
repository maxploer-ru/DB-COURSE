package service

import (
	"ZVideo/internal/domain/playlist/entity"
	"ZVideo/internal/domain/playlist/repository"
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrPlaylistNotFound       = errors.New("playlist not found")
	ErrVideoNotFound          = errors.New("video not found")
	ErrNotAuthorized          = errors.New("not authorized")
	ErrVideoAlreadyInPlaylist = errors.New("video already in playlist")
	ErrInvalidPosition        = errors.New("invalid position")
	ErrEmptyName              = errors.New("playlist name cannot be empty")
)

type ChannelChecker interface {
	IsOwner(ctx context.Context, channelID, userID int) (bool, error)
	Exists(ctx context.Context, channelID int) (bool, error)
}

type VideoChecker interface {
	Exists(ctx context.Context, videoID int) (bool, error)
}

type PlaylistService interface {
	CreatePlaylist(ctx context.Context, channelID, userID int, name, description string) (*entity.Playlist, error)
	GetPlaylist(ctx context.Context, playlistID int) (*entity.Playlist, error)
	UpdatePlaylist(ctx context.Context, playlistID, userID int, name, description *string) (*entity.Playlist, error)
	DeletePlaylist(ctx context.Context, playlistID, userID int) error
	AddVideoToPlaylist(ctx context.Context, playlistID, videoID, userID int, position *int) error
	RemoveVideoFromPlaylist(ctx context.Context, playlistID, videoID, userID int) error
	GetPlaylistVideos(ctx context.Context, playlistID int) ([]*entity.PlaylistItem, error)
	ReorderPlaylist(ctx context.Context, playlistID, userID int, videoIDs []int) error
}

type playlistService struct {
	playlistRepo     repository.PlaylistRepository
	playlistItemRepo repository.PlaylistItemRepository
	channelChecker   ChannelChecker
	videoChecker     VideoChecker
}

func NewPlaylistService(
	playlistRepo repository.PlaylistRepository,
	playlistItemRepo repository.PlaylistItemRepository,
	channelChecker ChannelChecker,
	videoChecker VideoChecker,
) PlaylistService {
	return &playlistService{
		playlistRepo:     playlistRepo,
		playlistItemRepo: playlistItemRepo,
		channelChecker:   channelChecker,
		videoChecker:     videoChecker,
	}
}

func (s *playlistService) CreatePlaylist(ctx context.Context, channelID, userID int, name, description string) (*entity.Playlist, error) {
	if name == "" {
		return nil, ErrEmptyName
	}

	exists, err := s.channelChecker.Exists(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("check channel exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("channel %d does not exist", channelID)
	}

	isOwner, err := s.channelChecker.IsOwner(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("check owner: %w", err)
	}
	if !isOwner {
		return nil, ErrNotAuthorized
	}

	playlist := &entity.Playlist{
		ChannelID:   channelID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}
	if err := s.playlistRepo.Create(ctx, playlist); err != nil {
		return nil, fmt.Errorf("create playlist: %w", err)
	}
	return playlist, nil
}

func (s *playlistService) GetPlaylist(ctx context.Context, playlistID int) (*entity.Playlist, error) {
	pl, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("get playlist: %w", err)
	}
	if pl == nil {
		return nil, ErrPlaylistNotFound
	}
	return pl, nil
}

func (s *playlistService) UpdatePlaylist(ctx context.Context, playlistID, userID int, name, description *string) (*entity.Playlist, error) {
	pl, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("get playlist: %w", err)
	}
	if pl == nil {
		return nil, ErrPlaylistNotFound
	}

	isOwner, err := s.channelChecker.IsOwner(ctx, pl.ChannelID, userID)
	if err != nil {
		return nil, fmt.Errorf("check owner: %w", err)
	}
	if !isOwner {
		return nil, ErrNotAuthorized
	}

	updated := false
	if name != nil && *name != pl.Name {
		if *name == "" {
			return nil, ErrEmptyName
		}
		pl.Name = *name
		updated = true
	}
	if description != nil && *description != pl.Description {
		pl.Description = *description
		updated = true
	}

	if updated {
		if err := s.playlistRepo.Update(ctx, pl); err != nil {
			return nil, fmt.Errorf("update playlist: %w", err)
		}
	}
	return pl, nil
}

func (s *playlistService) DeletePlaylist(ctx context.Context, playlistID, userID int) error {
	pl, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("get playlist: %w", err)
	}
	if pl == nil {
		return ErrPlaylistNotFound
	}

	isOwner, err := s.channelChecker.IsOwner(ctx, pl.ChannelID, userID)
	if err != nil {
		return fmt.Errorf("check owner: %w", err)
	}
	if !isOwner {
		return ErrNotAuthorized
	}

	if err := s.playlistRepo.Delete(ctx, playlistID); err != nil {
		return fmt.Errorf("delete playlist: %w", err)
	}
	return nil
}

func (s *playlistService) AddVideoToPlaylist(ctx context.Context, playlistID, videoID, userID int, position *int) error {
	pl, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("get playlist: %w", err)
	}
	if pl == nil {
		return ErrPlaylistNotFound
	}

	isOwner, err := s.channelChecker.IsOwner(ctx, pl.ChannelID, userID)
	if err != nil {
		return fmt.Errorf("check owner: %w", err)
	}
	if !isOwner {
		return ErrNotAuthorized
	}

	exists, err := s.videoChecker.Exists(ctx, videoID)
	if err != nil {
		return fmt.Errorf("check video exists: %w", err)
	}
	if !exists {
		return ErrVideoNotFound
	}

	existing, _ := s.playlistItemRepo.GetItem(ctx, playlistID, videoID)
	if existing != nil {
		return ErrVideoAlreadyInPlaylist
	}

	var pos int
	if position != nil {
		pos = *position
		if pos < 1 {
			return ErrInvalidPosition
		}

	} else {
		mx, err := s.playlistItemRepo.GetMaxPosition(ctx, playlistID)
		if err != nil {
			mx = 0
		}
		pos = mx + 1
	}

	item := &entity.PlaylistItem{
		PlaylistID: playlistID,
		VideoID:    videoID,
		Position:   pos,
		AddedAt:    time.Now(),
	}
	if err := s.playlistItemRepo.AddItem(ctx, item); err != nil {
		return fmt.Errorf("add item: %w", err)
	}
	return nil
}

func (s *playlistService) RemoveVideoFromPlaylist(ctx context.Context, playlistID, videoID, userID int) error {
	pl, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("get playlist: %w", err)
	}
	if pl == nil {
		return ErrPlaylistNotFound
	}

	isOwner, err := s.channelChecker.IsOwner(ctx, pl.ChannelID, userID)
	if err != nil {
		return fmt.Errorf("check owner: %w", err)
	}
	if !isOwner {
		return ErrNotAuthorized
	}

	if err := s.playlistItemRepo.RemoveItem(ctx, playlistID, videoID); err != nil {
		return fmt.Errorf("remove item: %w", err)
	}
	return nil
}

func (s *playlistService) GetPlaylistVideos(ctx context.Context, playlistID int) ([]*entity.PlaylistItem, error) {
	return s.playlistItemRepo.GetItemsByPlaylist(ctx, playlistID)
}

func (s *playlistService) ReorderPlaylist(ctx context.Context, playlistID, userID int, videoIDs []int) error {
	pl, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("get playlist: %w", err)
	}
	if pl == nil {
		return ErrPlaylistNotFound
	}

	isOwner, err := s.channelChecker.IsOwner(ctx, pl.ChannelID, userID)
	if err != nil {
		return fmt.Errorf("check owner: %w", err)
	}
	if !isOwner {
		return ErrNotAuthorized
	}

	return s.playlistItemRepo.ReorderItems(ctx, playlistID, videoIDs)
}
