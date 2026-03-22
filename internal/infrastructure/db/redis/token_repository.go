package redis

import (
	"ZVideo/internal/domain/auth/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type tokenRepository struct {
	client     *redis.Client
	refreshTTL time.Duration
}

func NewTokenRepository(client *redis.Client) repository.TokenRepository {
	return &tokenRepository{client: client}
}

func (r *tokenRepository) SaveRefreshToken(ctx context.Context, userID int, token string) error {
	pipe := r.client.Pipeline()

	key := fmt.Sprintf("refresh:%s", token)
	pipe.Set(ctx, key, userID, r.refreshTTL)

	userTokensKey := fmt.Sprintf("user:%d:tokens", userID)
	pipe.SAdd(ctx, userTokensKey, token)
	pipe.Expire(ctx, userTokensKey, r.refreshTTL)

	_, err := pipe.Exec(ctx)
	return err
}

func (r *tokenRepository) ValidateRefreshToken(ctx context.Context, token string) (int, error) {
	key := fmt.Sprintf("refresh:%s", token)
	userID, err := r.client.Get(ctx, key).Int()
	if errors.Is(err, redis.Nil) {
		return 0, fmt.Errorf("token not found")
	}
	return userID, err
}

func (r *tokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("refresh:%s", token)
	userID, err := r.client.Get(ctx, key).Int()
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()

	pipe.Del(ctx, key)

	userTokensKey := fmt.Sprintf("user:%d:tokens", userID)
	pipe.SRem(ctx, userTokensKey, token)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *tokenRepository) DeleteAllUserTokens(ctx context.Context, userID int) error {
	userTokensKey := fmt.Sprintf("user:%d:tokens", userID)
	tokens, err := r.client.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	for _, token := range tokens {
		refreshKey := fmt.Sprintf("refresh:%s", token)
		pipe.Del(ctx, refreshKey)
	}

	pipe.Del(ctx, userTokensKey)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *tokenRepository) BlacklistAccessToken(ctx context.Context, token string, ttl time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", token)
	return r.client.Set(ctx, key, "blacklisted", ttl).Err()
}

func (r *tokenRepository) IsAccessTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", token)
	_, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	return err == nil, err
}
