package main

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/config"
	"ZVideo/internal/infrastructure/db/postgres"
	pgrepository "ZVideo/internal/infrastructure/db/postgres/repository"
	"ZVideo/internal/infrastructure/logger"
	"ZVideo/internal/infrastructure/storage"
	"ZVideo/internal/worker"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.LoadConfig()
	appLogger, closeLog := logger.NewConfigured(cfg.Logging)
	defer closeLog()

	if err := run(cfg, appLogger); err != nil && !errors.Is(err, context.Canceled) {
		appLogger.ErrorContext(context.Background(), "storage worker stopped with error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(cfg *config.Config, appLogger domain.Logger) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get postgres sql connection: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	minioClient, _, err := storage.NewMinioClient(cfg.Minio)
	if err != nil {
		return fmt.Errorf("connect to minio: %w", err)
	}

	taskRepo := pgrepository.NewStorageDeleteTaskRepository(db)
	cleanupWorker := worker.NewStorageDeleteWorker(taskRepo, storage.NewMinioStorageService(minioClient, minioClient, cfg.Minio.Bucket), appLogger, 100, time.Minute)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	appLogger.InfoContext(ctx, "storage delete worker started", slog.Int("batch_size", 100), slog.Duration("interval", time.Minute))
	if err := cleanupWorker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("storage worker stopped with error: %w", err)
	}
	appLogger.InfoContext(ctx, "storage delete worker stopped")
	return nil
}
