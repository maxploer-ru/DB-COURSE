package storage

import (
	"ZVideo/internal/infrastructure/config"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioClient(cfg config.MinioConfig) (*minio.Client, *minio.Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := client.ListBuckets(ctx); err != nil {
		return nil, nil, fmt.Errorf("minio connection check failed: %w", err)
	}
	log.Printf("MinIO client connected to %s", cfg.Endpoint)

	externalEndpoint := cfg.ExternalEndpoint
	if externalEndpoint == "" {
		externalEndpoint = cfg.Endpoint
	}
	presignClient, err := minio.New(externalEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false,
		Region: "us-east-1",
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create presign minio client: %w", err)
	}

	return client, presignClient, nil
}

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
