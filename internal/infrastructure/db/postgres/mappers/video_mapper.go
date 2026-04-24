package mappers

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainVideo(dbVideo *models.Video) *domain.Video {
	if dbVideo == nil {
		return nil
	}
	return &domain.Video{
		ID:          dbVideo.ID,
		ChannelID:   dbVideo.ChannelID,
		ChannelName: dbVideo.Channel.Name,
		Title:       dbVideo.Title,
		Description: dbVideo.Description,
		Filepath:    dbVideo.Filepath,
		CreatedAt:   dbVideo.CreatedAt,
	}
}

func FromDomainVideo(domainVideo *domain.Video) *models.Video {
	if domainVideo == nil {
		return nil
	}
	return &models.Video{
		ID:          domainVideo.ID,
		ChannelID:   domainVideo.ChannelID,
		Title:       domainVideo.Title,
		Description: domainVideo.Description,
		Filepath:    domainVideo.Filepath,
		CreatedAt:   domainVideo.CreatedAt,
	}
}
