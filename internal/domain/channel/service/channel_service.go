package service

import (
	"ZVideo/internal/domain/channel/entity"
	"ZVideo/internal/domain/channel/repository"
	"context"
	"errors"
	"fmt"
)

var (
	ErrChannelNameExists = errors.New("channel name already exists")
)

type ChannelService interface {
	CreateChannel(ctx context.Context, userID int, name, description string) (*entity.Channel, error)
	GetChannel(ctx context.Context, id int) (*entity.Channel, error)
	UpdateChannel(ctx context.Context, channelID int, name, description *string) (*entity.Channel, error)
	DeleteChannel(ctx context.Context, channelID, userID int) error
}

type channelService struct {
	channelRepo repository.ChannelRepository
}

func NewChannelService(channelRepo repository.ChannelRepository) ChannelService {
	return &channelService{
		channelRepo: channelRepo,
	}
}

func (s *channelService) CreateChannel(ctx context.Context, userID int, name, description string) (*entity.Channel, error) {

	exists, err := s.channelRepo.ExistsByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("check name: %w", err)
	}
	if exists {
		return nil, ErrChannelNameExists
	}

	channel := &entity.Channel{
		UserID:      userID,
		Name:        name,
		Description: description,
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}
	return channel, nil
}

func (s *channelService) GetChannel(ctx context.Context, id int) (*entity.Channel, error) {
	ch, err := s.channelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}
	if ch == nil {
		return nil, ErrChannelNotFound
	}
	return ch, nil
}

func (s *channelService) UpdateChannel(ctx context.Context, channelID int, name, description *string) (*entity.Channel, error) {
	ch, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}
	if ch == nil {
		return nil, ErrChannelNotFound
	}

	if name != nil && *name != ch.Name {
		exists, err := s.channelRepo.ExistsByName(ctx, *name)
		if err != nil {
			return nil, fmt.Errorf("check name: %w", err)
		}
		if exists {
			return nil, ErrChannelNameExists
		}
		ch.Name = *name
	}
	if description != nil {
		ch.Description = *description
	}

	if err := s.channelRepo.Update(ctx, ch); err != nil {
		return nil, fmt.Errorf("update channel: %w", err)
	}
	return ch, nil
}

func (s *channelService) DeleteChannel(ctx context.Context, channelID, userID int) error {
	ch, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("get channel: %w", err)
	}
	if ch == nil {
		return ErrChannelNotFound
	}

	if ch.UserID != userID {
		return errors.New("not authorized to delete this channel")
	}
	if err := s.channelRepo.Delete(ctx, channelID); err != nil {
		return fmt.Errorf("delete channel: %w", err)
	}
	return nil
}
