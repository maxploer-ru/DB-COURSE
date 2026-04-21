package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"fmt"
	"log/slog"
)

type ChannelService interface {
	CreateChannel(ctx context.Context, userID int, name, description string) (*domain.Channel, error)
	GetChannel(ctx context.Context, id int) (*domain.Channel, error)
	GetChannelByName(ctx context.Context, name string) (*domain.Channel, error)
	GetChannelByUserID(ctx context.Context, userID int) (*domain.Channel, error)
	UpdateChannel(ctx context.Context, channelID, userID int, name, description *string) (*domain.Channel, error)
	DeleteChannel(ctx context.Context, channelID, userID int) error
	Exists(ctx context.Context, channelID int) (bool, error)
	IsOwner(ctx context.Context, channelID, userID int) (bool, error)
}

type channelVideoFilepathRepository interface {
	ListFilepathsByChannel(ctx context.Context, channelID int) ([]string, error)
}

type channelService struct {
	channelRepo repository.ChannelRepository
	videoRepo   channelVideoFilepathRepository
	storageSvc  StorageService
}

func NewChannelService(
	channelRepo repository.ChannelRepository,
	videoRepo channelVideoFilepathRepository,
	storageSvc StorageService,
) ChannelService {
	return &channelService{
		channelRepo: channelRepo,
		videoRepo:   videoRepo,
		storageSvc:  storageSvc,
	}
}

func (s *channelService) CreateChannel(ctx context.Context, userID int, name, description string) (*domain.Channel, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "ChannelService"),
		slog.String("operation", "CreateChannel"),
		slog.Int("user_id", userID),
		slog.String("channel_name", name),
	)

	logger.DebugContext(ctx, "Checking if user already has a channel")
	existing, err := s.channelRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check existing channel", slog.String("error", err.Error()))
		return nil, fmt.Errorf("check existence of channel failed: %w", err)
	}
	if existing != nil {
		logger.WarnContext(ctx, "User already has a channel", slog.Int("existing_channel_id", existing.ID))
		return nil, domain.ErrChannelAlreadyExists
	}

	logger.DebugContext(ctx, "Checking channel name uniqueness")
	exists, err := s.channelRepo.ExistsByName(ctx, name)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel name", slog.String("error", err.Error()))
		return nil, fmt.Errorf("check name failed: %w", err)
	}
	if exists {
		logger.WarnContext(ctx, "Channel name already taken")
		return nil, domain.ErrChannelNameAlreadyExists
	}

	channel := &domain.Channel{
		UserID:      userID,
		Name:        name,
		Description: description,
	}

	logger.DebugContext(ctx, "Creating channel in repository")
	if err := s.channelRepo.Create(ctx, channel); err != nil {
		logger.ErrorContext(ctx, "Failed to create channel in repository", slog.String("error", err.Error()))
		return nil, fmt.Errorf("create channel failed: %w", err)
	}

	logger.InfoContext(ctx, "Channel created successfully", slog.Int("channel_id", channel.ID))
	return channel, nil
}

func (s *channelService) GetChannel(ctx context.Context, id int) (*domain.Channel, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "ChannelService"),
		slog.String("operation", "GetChannel"),
		slog.Int("channel_id", id),
	)

	logger.DebugContext(ctx, "Fetching channel by ID")
	ch, err := s.channelRepo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get channel", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get channel failed: %w", err)
	}
	if ch == nil {
		logger.WarnContext(ctx, "Channel not found")
		return nil, domain.ErrChannelNotFound
	}

	logger.DebugContext(ctx, "Channel retrieved successfully")
	return ch, nil
}

func (s *channelService) GetChannelByName(ctx context.Context, name string) (*domain.Channel, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "ChannelService"),
		slog.String("operation", "GetChannelByName"),
		slog.String("channel_name", name),
	)

	logger.DebugContext(ctx, "Fetching channel by name")
	ch, err := s.channelRepo.GetByName(ctx, name)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get channel by name", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get channel failed: %w", err)
	}
	if ch == nil {
		logger.WarnContext(ctx, "Channel not found")
		return nil, domain.ErrChannelNotFound
	}

	logger.DebugContext(ctx, "Channel retrieved successfully", slog.Int("channel_id", ch.ID))
	return ch, nil
}

func (s *channelService) GetChannelByUserID(ctx context.Context, userID int) (*domain.Channel, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "ChannelService"),
		slog.String("operation", "GetChannelByUserID"),
		slog.Int("user_id", userID),
	)

	logger.DebugContext(ctx, "Fetching channel by user ID")
	ch, err := s.channelRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get channel by user ID", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get channel failed: %w", err)
	}
	if ch == nil {
		logger.WarnContext(ctx, "Channel not found for user")
		return nil, domain.ErrChannelNotFound
	}

	logger.DebugContext(ctx, "Channel retrieved successfully", slog.Int("channel_id", ch.ID))
	return ch, nil
}

func (s *channelService) UpdateChannel(ctx context.Context, channelID, userID int, name, description *string) (*domain.Channel, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "ChannelService"),
		slog.String("operation", "UpdateChannel"),
		slog.Int("channel_id", channelID),
		slog.Int("user_id", userID),
	)

	logger.DebugContext(ctx, "Fetching channel for update")
	ch, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get channel", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get channel failed: %w", err)
	}
	if ch == nil {
		logger.WarnContext(ctx, "Channel not found")
		return nil, domain.ErrChannelNotFound
	}

	if ch.UserID != userID {
		logger.WarnContext(ctx, "User is not the channel owner")
		return nil, domain.ErrForbidden
	}

	updated := false
	if name != nil && *name != ch.Name {
		logger.DebugContext(ctx, "Checking new channel name uniqueness", slog.String("new_name", *name))
		exists, err := s.channelRepo.ExistsByName(ctx, *name)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to check channel name", slog.String("error", err.Error()))
			return nil, fmt.Errorf("check name failed: %w", err)
		}
		if exists {
			logger.WarnContext(ctx, "New channel name already taken")
			return nil, domain.ErrChannelNameAlreadyExists
		}
		ch.Name = *name
		updated = true
	}
	if description != nil {
		ch.Description = *description
		updated = true
	}

	if !updated {
		logger.DebugContext(ctx, "No changes to update")
		return ch, nil
	}

	logger.DebugContext(ctx, "Updating channel in repository")
	if err := s.channelRepo.Update(ctx, ch); err != nil {
		logger.ErrorContext(ctx, "Failed to update channel", slog.String("error", err.Error()))
		return nil, fmt.Errorf("update channel failed: %w", err)
	}

	logger.InfoContext(ctx, "Channel updated successfully")
	return ch, nil
}

func (s *channelService) DeleteChannel(ctx context.Context, channelID, userID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "ChannelService"),
		slog.String("operation", "DeleteChannel"),
		slog.Int("channel_id", channelID),
		slog.Int("user_id", userID),
	)

	logger.DebugContext(ctx, "Fetching channel for deletion")
	ch, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get channel", slog.String("error", err.Error()))
		return fmt.Errorf("get channel failed: %w", err)
	}
	if ch == nil {
		logger.WarnContext(ctx, "Channel not found")
		return domain.ErrChannelNotFound
	}

	if ch.UserID != userID {
		logger.WarnContext(ctx, "User is not the channel owner")
		return domain.ErrForbidden
	}

	logger.DebugContext(ctx, "Fetching channel video filepaths for storage cleanup")
	filepaths, err := s.videoRepo.ListFilepathsByChannel(ctx, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to fetch channel videos for storage cleanup", slog.String("error", err.Error()))
		return fmt.Errorf("list channel video filepaths failed: %w", err)
	}

	if len(filepaths) > 0 {
		logger.DebugContext(ctx, "Deleting channel videos from storage", slog.Int("video_count", len(filepaths)))
		failedDeletes := 0
		for _, filepath := range filepaths {
			if filepath == "" {
				continue
			}
			if err := s.storageSvc.DeleteObject(ctx, filepath); err != nil {
				failedDeletes++
				logger.WarnContext(ctx, "Failed to delete video file from storage",
					slog.String("filepath", filepath),
					slog.String("error", err.Error()),
				)
			}
		}
		if failedDeletes > 0 {
			logger.ErrorContext(ctx, "Channel deletion cancelled: some video files were not deleted from storage",
				slog.Int("failed_count", failedDeletes),
			)
			return fmt.Errorf("delete channel video files from storage failed: %d object(s)", failedDeletes)
		}
	}

	logger.DebugContext(ctx, "Deleting channel from repository")
	if err := s.channelRepo.Delete(ctx, channelID); err != nil {
		logger.ErrorContext(ctx, "Failed to delete channel", slog.String("error", err.Error()))
		return fmt.Errorf("delete channel failed: %w", err)
	}

	logger.InfoContext(ctx, "Channel deleted successfully")
	return nil
}

func (s *channelService) Exists(ctx context.Context, channelID int) (bool, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "ChannelService"),
		slog.String("operation", "Exists"),
		slog.Int("channel_id", channelID),
	)

	logger.DebugContext(ctx, "Checking channel existence")
	ch, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check channel existence", slog.String("error", err.Error()))
		return false, fmt.Errorf("check channel exists failed: %w", err)
	}
	exists := ch != nil
	logger.DebugContext(ctx, "Channel existence checked", slog.Bool("exists", exists))
	return exists, nil
}

func (s *channelService) IsOwner(ctx context.Context, channelID, userID int) (bool, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "ChannelService"),
		slog.String("operation", "IsOwner"),
		slog.Int("channel_id", channelID),
		slog.Int("user_id", userID),
	)

	logger.DebugContext(ctx, "Checking channel ownership")
	ch, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get channel", slog.String("error", err.Error()))
		return false, fmt.Errorf("check channel owner failed: %w", err)
	}
	if ch == nil {
		logger.WarnContext(ctx, "Channel not found")
		return false, domain.ErrChannelNotFound
	}
	isOwner := ch.UserID == userID
	logger.DebugContext(ctx, "Ownership checked", slog.Bool("is_owner", isOwner))
	return isOwner, nil
}
