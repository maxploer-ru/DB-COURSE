package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinioStorageClient struct {
	client     *minio.Client
	bucketName string
}

func NewMinioStorageService(client *minio.Client, bucketName string) *MinioStorageClient {
	return &MinioStorageClient{
		client:     client,
		bucketName: bucketName,
	}
}

func (s *MinioStorageClient) GenerateUploadPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedPutObject(ctx, s.bucketName, key, expiry)
	if err != nil {
		return "", fmt.Errorf("presigned put object: %w", err)
	}
	return url.String(), nil
}

func (s *MinioStorageClient) GenerateAccessPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucketName, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presigned get object: %w", err)
	}
	return url.String(), nil
}

func (s *MinioStorageClient) DeleteObject(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("remove object: %w", err)
	}
	return nil
}
