package mappers

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainVideo(dbVideo *models.Video) *domain.Video {
	if dbVideo == nil {
		return nil
	}

	domainVideo := &domain.Video{
		ID:               dbVideo.ID,
		ChannelID:        dbVideo.ChannelID,
		Title:            dbVideo.Title,
		Description:      dbVideo.Description,
		Filepath:         dbVideo.Filepath,
		OriginalFilename: dbVideo.OriginalFilename,
		Status:           domain.VideoStatus(dbVideo.Status),
		CreatedAt:        dbVideo.CreatedAt,
	}

	if dbVideo.Channel.ID != 0 || dbVideo.Channel.Name != "" {
		domainVideo.ChannelName = dbVideo.Channel.Name
	}

	return domainVideo
}

func FromDomainVideo(domainVideo *domain.Video) *models.Video {
	if domainVideo == nil {
		return nil
	}
	return &models.Video{
		ID:               domainVideo.ID,
		ChannelID:        domainVideo.ChannelID,
		Title:            domainVideo.Title,
		Description:      domainVideo.Description,
		Filepath:         domainVideo.Filepath,
		OriginalFilename: domainVideo.OriginalFilename,
		Status:           string(domainVideo.Status),
		CreatedAt:        domainVideo.CreatedAt,
	}
}
