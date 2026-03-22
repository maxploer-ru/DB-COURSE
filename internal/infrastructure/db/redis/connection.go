package redis

import (
	"ZVideo/internal/infrastructure/config"
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

func NewClient(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
		PoolSize: 10,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Failed to close Redis client after ping failure: %v", closeErr)
		}
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Connected to Redis")
	return client, nil
}
