package domain

import "time"

type Subscription struct {
	UserID         int
	ChannelID      int
	NewVideosCount int
	SubscribedAt   time.Time
}
