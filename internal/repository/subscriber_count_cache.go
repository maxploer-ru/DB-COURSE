package repository

import "context"

type SubscriberCounter interface {
	Increment(ctx context.Context, channelID int) error
	Decrement(ctx context.Context, channelID int) error
	Get(ctx context.Context, channelID int) (count int, hit bool, err error)
	LoadAll(ctx context.Context) (map[int]int, error)
	Set(ctx context.Context, channelID int, count int) error
}
