package storage

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinioStorage struct {
	client *minio.Client
}

func NewMinioStorage(client *minio.Client) *MinioStorage {
	return &MinioStorage{
		client: client,
	}
}

func (m *MinioStorage) UploadFile(ctx context.Context, bucket, fileName string, reader io.Reader, size int64, contentType string) (string, error) {
	exists, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return "", err
	}
	if !exists {
		err = m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return "", err
		}
	}

	_, err = m.client.PutObject(ctx, bucket, fileName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return "/" + bucket + "/" + fileName, nil
}

func (m *MinioStorage) GenerateUploadURL(ctx context.Context, bucket, fileName string, expires time.Duration) (string, error) {
	url, err := m.client.PresignedPutObject(ctx, bucket, fileName, expires)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
