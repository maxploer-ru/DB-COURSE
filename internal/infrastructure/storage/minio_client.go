package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

// NewMinioClient создаёт и проверяет подключение к MinIO.
func NewMinioClient(cfg MinioConfig) (*minio.Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Healthcheck соединения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := client.ListBuckets(ctx); err != nil {
		return nil, fmt.Errorf("minio connection check failed: %w", err)
	}

	log.Printf("MinIO client connected to %s", cfg.Endpoint)
	return client, nil
}

// EnsureBucketExists проверяет существование бакета и создаёт его при необходимости.
func EnsureBucketExists(ctx context.Context, client *minio.Client, bucketName string) error {
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("check bucket existence failed: %w", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("create bucket failed: %w", err)
		}
		log.Printf("Bucket '%s' created", bucketName)
	}
	return nil
}
