package worker

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type ObjectDeleter interface {
	DeleteObject(ctx context.Context, key string) error
}

type StorageDeleteWorker struct {
	tasks      repository.StorageDeleteTaskRepository
	deleter    ObjectDeleter
	logger     domain.Logger
	batchSize  int
	interval   time.Duration
	maxBackoff time.Duration
}

func NewStorageDeleteWorker(
	tasks repository.StorageDeleteTaskRepository,
	deleter ObjectDeleter,
	logger domain.Logger,
	batchSize int,
	interval time.Duration,
) *StorageDeleteWorker {
	if logger == nil {
		logger = domain.GetLogger(context.Background())
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	if interval <= 0 {
		interval = time.Minute
	}
	logger = logger.With(slog.String("component", "storage_delete_worker"))
	return &StorageDeleteWorker{
		tasks:      tasks,
		deleter:    deleter,
		logger:     logger,
		batchSize:  batchSize,
		interval:   interval,
		maxBackoff: time.Hour,
	}
}

func (w *StorageDeleteWorker) Run(ctx context.Context) error {
	if err := w.ProcessOnce(ctx); err != nil {
		w.logger.ErrorContext(ctx, "initial storage cleanup failed", slog.String("error", err.Error()))
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.ProcessOnce(ctx); err != nil {
				w.logger.ErrorContext(ctx, "storage cleanup failed", slog.String("error", err.Error()))
			}
		}
	}
}

func (w *StorageDeleteWorker) ProcessOnce(ctx context.Context) error {
	tasks, err := w.tasks.ClaimBatch(ctx, w.batchSize)
	if err != nil {
		w.logger.ErrorContext(ctx, "failed to claim storage cleanup tasks", slog.Int("batch_size", w.batchSize), slog.Any("error", err))
		return err
	}
	w.logger.DebugContext(ctx, "storage cleanup tasks claimed", slog.Int("count", len(tasks)))
	for _, task := range tasks {
		if err := w.deleter.DeleteObject(ctx, task.Filepath); err != nil {
			backoff := w.retryBackoff(task.Attempts)
			if markErr := w.tasks.MarkFailed(ctx, task.ID, time.Now().UTC().Add(backoff), err.Error()); markErr != nil {
				w.logger.ErrorContext(ctx, "failed to persist storage cleanup retry", slog.Int("task_id", task.ID), slog.String("error", markErr.Error()))
			}
			w.logger.WarnContext(ctx, "storage object deletion failed",
				slog.Int("task_id", task.ID),
				slog.String("filepath", task.Filepath),
				slog.String("error", err.Error()),
				slog.Duration("retry_in", backoff),
			)
			continue
		}
		if err := w.tasks.MarkDone(ctx, task.ID); err != nil {
			w.logger.ErrorContext(ctx, "failed to mark storage cleanup task done", slog.Int("task_id", task.ID), slog.Any("error", err))
			return fmt.Errorf("mark completed storage task %d: %w", task.ID, err)
		}
	}
	return nil
}

func (w *StorageDeleteWorker) retryBackoff(attempts int) time.Duration {
	if attempts < 0 {
		attempts = 0
	}
	// The first retry is one minute; the delay is capped to avoid hot-looping
	// on a permanently unavailable storage backend.
	backoff := time.Minute * time.Duration(1<<min(attempts, 10))
	if backoff > w.maxBackoff {
		return w.maxBackoff
	}
	return backoff
}
