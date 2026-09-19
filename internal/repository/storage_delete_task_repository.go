package repository

import (
	"ZVideo/internal/domain"
	"context"
	"time"
)

type StorageDeleteTaskRepository interface {
	ClaimBatch(ctx context.Context, limit int) ([]*domain.StorageDeleteTask, error)
	MarkDone(ctx context.Context, id int) error
	MarkFailed(ctx context.Context, id int, nextAttemptAt time.Time, lastError string) error
}
