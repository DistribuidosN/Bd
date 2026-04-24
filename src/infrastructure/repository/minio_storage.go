package repository

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
	// Mantenemos GetURL con 24h para compatibilidad actual
	return r.GetPresignedURL(ctx, fileName, time.Hour*24)
}

func (r *minioStorageRepository) GetPresignedURL(ctx context.Context, fileName string, duration time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := r.client.PresignedGetObject(ctx, r.bucket, fileName, duration, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned url: %w", err)
	}
	return presignedURL.String(), nil
}

func (r *minioStorageRepository) Download(ctx context.Context, fileName string) ([]byte, error) {
	object, err := r.client.GetObject(ctx, r.bucket, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from minio: %w", err)
	}
	defer object.Close()

	return io.ReadAll(object)
}
