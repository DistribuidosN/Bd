package repository

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"time"

	"enfok_bd/src/domain/ports/driven"

	"github.com/minio/minio-go/v7"
)

type minioStorageRepository struct {
	client *minio.Client
	bucket string
}

func NewMinioStorageRepository(client *minio.Client, bucket string) driven.StorageRepository {
	return &minioStorageRepository{client: client, bucket: bucket}
}

func (r *minioStorageRepository) Upload(ctx context.Context, fileName string, content []byte) (string, error) {
	_, err := r.client.PutObject(ctx, r.bucket, fileName, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to minio: %w", err)
	}
	return fileName, nil
}

func (r *minioStorageRepository) GetURL(ctx context.Context, fileName string) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := r.client.PresignedGetObject(ctx, r.bucket, fileName, time.Second*24*60*60, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned url: %w", err)
	}
	return presignedURL.String(), nil
}
