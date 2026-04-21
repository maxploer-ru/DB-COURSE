package mappers

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainChannel(model *models.Channel) *domain.Channel {
	if model == nil {
		return nil
	}
	return &domain.Channel{
		ID:          model.ID,
		Name:        model.Name,
		CreatedAt:   model.CreatedAt,
		UserID:      model.UserID,
		Description: model.Description,
	}
}

func FromDomainChannel(model *domain.Channel) *models.Channel {
	if model == nil {
		return nil
	}
	return &models.Channel{
		ID:          model.ID,
		Name:        model.Name,
		CreatedAt:   model.CreatedAt,
		UserID:      model.UserID,
		Description: model.Description,
	}
}

func ToDomainChannelList(models []*models.Channel) []*domain.Channel {
	channels := make([]*domain.Channel, len(models))
	for i, model := range models {
		channels[i] = ToDomainChannel(model)
	}
	return channels
}

func FromDomainChannelList(channels []*domain.Channel) []*models.Channel {
	channelsModels := make([]*models.Channel, len(channels))
	for i, channel := range channels {
		channelsModels[i] = FromDomainChannel(channel)
	}
	return channelsModels
}
