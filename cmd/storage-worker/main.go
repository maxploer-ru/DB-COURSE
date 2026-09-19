package main

import (
	"ZVideo/internal/infrastructure/config"
	"ZVideo/internal/infrastructure/db/postgres"
	pgrepository "ZVideo/internal/infrastructure/db/postgres/repository"
	"ZVideo/internal/infrastructure/storage"
	"ZVideo/internal/worker"
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil && !errors.Is(err, context.Canceled) {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.LoadConfig()
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
	cleanupWorker := worker.NewStorageDeleteWorker(taskRepo, storage.NewMinioStorageService(minioClient, minioClient, cfg.Minio.Bucket), slog.Default(), 100, time.Minute)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := cleanupWorker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("storage worker stopped with error: %w", err)
	}
	return nil
}
