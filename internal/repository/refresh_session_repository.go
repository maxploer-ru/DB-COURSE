package repository

import (
	"context"
	"time"
)

type RefreshSessionRepository interface {
	Save(ctx context.Context, tokenID string, userID int, expiresAt time.Time) error
	GetUserID(ctx context.Context, tokenID string) (userID int, found bool, err error)
	Rotate(ctx context.Context, oldTokenID, newTokenID string, userID int, expiresAt time.Time) (rotated bool, err error)
	Delete(ctx context.Context, tokenID string) error
}
