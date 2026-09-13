package builder

import (
	"ZVideo/internal/domain"
	"time"
)

type ChannelBuilder struct {
	channel *domain.Channel
}

func NewChannelBuilder() *ChannelBuilder {
	return &ChannelBuilder{
		channel: &domain.Channel{
			ID:            1,
			UserID:        1,
			OwnerUsername: "test_user",
			Name:          "test_channel",
			Description:   "Test Description",
			CreatedAt:     time.Now(),
		},
	}
}

func (b *ChannelBuilder) WithID(id int) *ChannelBuilder {
	b.channel.ID = id
	return b
}

func (b *ChannelBuilder) WithUserID(userID int) *ChannelBuilder {
	b.channel.UserID = userID
	return b
}

func (b *ChannelBuilder) WithName(name string) *ChannelBuilder {
	b.channel.Name = name
	return b
}

func (b *ChannelBuilder) Build() *domain.Channel {
	return b.channel
}
