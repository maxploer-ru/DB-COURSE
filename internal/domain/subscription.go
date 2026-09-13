package domain

import "time"

type Subscription struct {
	UserID         int
	ChannelID      int
	ChannelName    string
	NewVideosCount int
	SubscribedAt   time.Time
}
