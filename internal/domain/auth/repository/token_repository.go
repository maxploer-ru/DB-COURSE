package repository

import (
	"context"
	"time"
)

type TokenRepository interface {
	SaveRefreshToken(ctx context.Context, userID int, token string) error

	ValidateRefreshToken(ctx context.Context, token string) (int, error)

	DeleteRefreshToken(ctx context.Context, token string) error

	DeleteAllUserTokens(ctx context.Context, userID int) error

	BlacklistAccessToken(ctx context.Context, token string, ttl time.Duration) error
	IsAccessTokenBlacklisted(ctx context.Context, token string) (bool, error)
}
