package entity

import "time"

type Viewing struct {
	ID        int
	UserID    int
	VideoID   int
	WatchedAt time.Time
}
