package repository

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"enfok_bd/src/domain/ports/driven"
	"enfok_bd/src/utils"

	"github.com/minio/minio-go/v7"
)

type minioStorageRepository struct {
	internalClient *minio.Client
	externalClient *minio.Client
	bucket         string
}

func NewMinioStorageRepository(internalClient *minio.Client, externalClient *minio.Client, bucket string) driven.StorageRepository {
	return &minioStorageRepository{
		internalClient: internalClient,
		externalClient: externalClient,
		bucket:         bucket,
	}
}

func (r *minioStorageRepository) Upload(ctx context.Context, fileName string, content []byte) (string, error) {
	_, err := r.internalClient.PutObject(ctx, r.bucket, fileName, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to minio: %w", err)
	}
	return fileName, nil
}

func (r *minioStorageRepository) GetURL(ctx context.Context, fileName string) (string, error) {
	return r.GetPresignedURL(ctx, fileName)
}

func (r *minioStorageRepository) GetPresignedURL(ctx context.Context, fileName string) (string, error) {
	ip := utils.GetLocalIP()
	return fmt.Sprintf("http://%s/%s/%s", ip, r.bucket, fileName), nil
}

func (r *minioStorageRepository) Download(ctx context.Context, fileName string) ([]byte, error) {
	object, err := r.internalClient.GetObject(ctx, r.bucket, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from minio: %w", err)
	}
	defer object.Close()

	return io.ReadAll(object)
}

func (r *minioStorageRepository) UploadStream(ctx context.Context, bucketName string, fileName string, reader io.Reader, size int64, contentType string) error {
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}
	if size == -1 {
		opts.PartSize = 10 * 1024 * 1024 // 10MB
	}
	_, err := r.internalClient.PutObject(ctx, bucketName, fileName, reader, size, opts)
	if err != nil {
		return fmt.Errorf("failed to upload stream to minio bucket %s: %w", bucketName, err)
	}
	return nil
}

func (r *minioStorageRepository) GetExportPresignedURL(ctx context.Context, fileName string) (string, error) {
	ip := utils.GetLocalIP()
	return fmt.Sprintf("http://%s/exports/%s", ip, fileName), nil
}

func (r *minioStorageRepository) DeleteOldExports(ctx context.Context, duration time.Duration) error {
	bucketName := "exports"
	cutoff := time.Now().Add(-duration)

	opts := minio.ListObjectsOptions{
		Recursive: true,
	}

	for object := range r.internalClient.ListObjects(ctx, bucketName, opts) {
		if object.Err != nil {
			continue // skip errors
		}
		if object.LastModified.Before(cutoff) {
			_ = r.internalClient.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{})
		}
	}
	return nil
}
