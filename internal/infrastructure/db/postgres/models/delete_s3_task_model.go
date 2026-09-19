package models

import "time"

type DeleteS3Task struct {
	ID            int       `gorm:"primaryKey"`
	Filepath      string    `gorm:"not null"`
	IsDone        bool      `gorm:"not null"`
	Attempts      int       `gorm:"not null"`
	NextAttemptAt time.Time `gorm:"not null"`
	LastError     *string
	LockedAt      *time.Time
	ProcessedAt   *time.Time
	CreatedAt     time.Time `gorm:"not null"`
}

func (DeleteS3Task) TableName() string {
	return "delete_s3_tasks"
}
