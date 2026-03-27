package entity

import "time"

type Subscription struct {
	UserID       int
	ChannelID    int
	SubscribedAt time.Time
}
