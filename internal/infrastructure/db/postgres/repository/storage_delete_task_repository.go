package repository

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type StorageDeleteTaskRepository struct {
	db *gorm.DB
}

func NewStorageDeleteTaskRepository(db *gorm.DB) *StorageDeleteTaskRepository {
	return &StorageDeleteTaskRepository{db: db}
}

func (r *StorageDeleteTaskRepository) ClaimBatch(ctx context.Context, limit int) ([]*domain.StorageDeleteTask, error) {
	if limit <= 0 {
		return nil, nil
	}

	var tasks []*domain.StorageDeleteTask
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rows []models.DeleteS3Task
		query := `
SELECT id, filepath, attempts, next_attempt_at
FROM delete_s3_tasks
WHERE is_done = FALSE
  AND next_attempt_at <= NOW()
  AND (locked_at IS NULL OR locked_at < NOW() - INTERVAL '5 minutes')
ORDER BY id
FOR UPDATE SKIP LOCKED
LIMIT ?`
		if err := tx.Raw(query, limit).Scan(&rows).Error; err != nil {
			return fmt.Errorf("claim delete tasks: %w", err)
		}

		if len(rows) == 0 {
			tasks = []*domain.StorageDeleteTask{}
			return nil
		}

		now := time.Now().UTC()
		tasks = make([]*domain.StorageDeleteTask, 0, len(rows))
		for _, row := range rows {
			result := tx.Model(&models.DeleteS3Task{}).
				Where("id = ?", row.ID).
				Updates(map[string]any{"locked_at": now})
			if result.Error != nil {
				return fmt.Errorf("lock delete task %d: %w", row.ID, result.Error)
			}
			tasks = append(tasks, &domain.StorageDeleteTask{
				ID:            row.ID,
				Filepath:      row.Filepath,
				Attempts:      row.Attempts,
				NextAttemptAt: row.NextAttemptAt,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *StorageDeleteTaskRepository) MarkDone(ctx context.Context, id int) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&models.DeleteS3Task{}).
		Where("id = ? AND is_done = FALSE", id).
		Updates(map[string]any{
			"is_done":      true,
			"processed_at": now,
			"locked_at":    nil,
			"last_error":   nil,
		})
	if result.Error != nil {
		return fmt.Errorf("mark delete task %d done: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete task %d not found or already completed", id)
	}
	return nil
}

func (r *StorageDeleteTaskRepository) MarkFailed(ctx context.Context, id int, nextAttemptAt time.Time, lastError string) error {
	result := r.db.WithContext(ctx).Model(&models.DeleteS3Task{}).
		Where("id = ? AND is_done = FALSE", id).
		Updates(map[string]any{
			"attempts":        gorm.Expr("attempts + 1"),
			"next_attempt_at": nextAttemptAt,
			"last_error":      lastError,
			"locked_at":       nil,
		})
	if result.Error != nil {
		return fmt.Errorf("mark delete task %d failed: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete task %d not found or already completed", id)
	}
	return nil
}
