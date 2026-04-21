package repository

import "context"

type CommentStatsCache interface {
	IncrLikes(ctx context.Context, commentID int) error
	DecrLikes(ctx context.Context, commentID int) error
	IncrDislikes(ctx context.Context, commentID int) error
	DecrDislikes(ctx context.Context, commentID int) error
	GetStats(ctx context.Context, commentID int) (likes, dislikes int64, hit bool, err error)
	SetStats(ctx context.Context, commentID int, likes, dislikes int64) error
}
