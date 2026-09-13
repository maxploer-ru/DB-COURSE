package builder

import (
	"ZVideo/internal/domain"
	"time"
)

type VideoBuilder struct {
	video *domain.Video
}

func NewVideoBuilder() *VideoBuilder {
	return &VideoBuilder{
		video: &domain.Video{
			ID:               1,
			ChannelID:        1,
			ChannelName:      "test_channel",
			Title:            "test_video",
			Description:      "description",
			Filepath:         "videos/1/test.mp4",
			OriginalFilename: "test.mp4",
			Status:           domain.VideoStatusReady,
			CreatedAt:        time.Now(),
		},
	}
}

func (b *VideoBuilder) WithID(id int) *VideoBuilder {
	b.video.ID = id
	return b
}

func (b *VideoBuilder) WithChannelID(channelID int) *VideoBuilder {
	b.video.ChannelID = channelID
	return b
}

func (b *VideoBuilder) Pending() *VideoBuilder {
	b.video.Status = domain.VideoStatusPending
	return b
}

func (b *VideoBuilder) Build() *domain.Video {
	return b.video
}
