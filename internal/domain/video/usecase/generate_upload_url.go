package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"path/filepath"
	"time"
)

type URLStorageProvider interface {
	GenerateUploadURL(ctx context.Context, bucket, fileName string, expires time.Duration) (string, error)
}

type GenerateUploadURLUseCase struct {
	Storage    URLStorageProvider
	BucketName string
}

func NewGenerateUploadURLUseCase(storage URLStorageProvider, bucketName string) *GenerateUploadURLUseCase {
	return &GenerateUploadURLUseCase{
		Storage:    storage,
		BucketName: bucketName,
	}
}

func (uc *GenerateUploadURLUseCase) Execute(ctx context.Context, originalFilename string) (string, string, error) {

	b := make([]byte, 16)
	_, _ = rand.Read(b)
	ext := filepath.Ext(originalFilename)
	fileName := fmt.Sprintf("%x%s", b, ext)

	url, err := uc.Storage.GenerateUploadURL(ctx, uc.BucketName, fileName, 24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return url, "/" + uc.BucketName + "/" + fileName, nil
}
