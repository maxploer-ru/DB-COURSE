package mappers

import (
	"ZVideo/internal/delivery/handlers/dto"
	"ZVideo/internal/domain"
	"time"
)

func ToVideoResponse(video *domain.Video, stats *domain.VideoStats) *dto.VideoResponse {
	if video == nil || stats == nil {
		return nil
	}
	return &dto.VideoResponse{
		ID:          video.ID,
		ChannelID:   video.ChannelID,
		Title:       video.Title,
		Description: video.Description,
		Views:       stats.Views,
		Likes:       stats.Likes,
		Dislikes:    stats.Dislikes,
		Comments:    stats.Comments,
		CreatedAt:   video.CreatedAt.Format(time.RFC3339),
	}
}

func ToVideoListResponse(videos []*domain.Video, stats []*domain.VideoStats) []*dto.VideoResponse {
	if videos == nil || stats == nil {
		return nil
	}
	resp := make([]*dto.VideoResponse, len(videos))
	for i, v := range videos {
		resp[i] = ToVideoResponse(v, stats[i])
	}
	return resp
}
