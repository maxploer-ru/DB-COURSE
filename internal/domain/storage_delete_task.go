package domain

import "time"

// StorageDeleteTask is an outbox item for deleting an object outside the
// database transaction that removed its owning row.
type StorageDeleteTask struct {
	ID            int
	Filepath      string
	Attempts      int
	NextAttemptAt time.Time
}
