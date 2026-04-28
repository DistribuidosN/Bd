package driven

import (
	"context"
	"io"
	"time"
)

type StorageRepository interface {
	Upload(ctx context.Context, fileName string, content []byte) (string, error)
	UploadStream(ctx context.Context, bucketName string, fileName string, reader io.Reader, size int64, contentType string) error
	GetURL(ctx context.Context, fileName string) (string, error)
	GetPresignedURL(ctx context.Context, fileName string, duration time.Duration) (string, error)
	GetExportPresignedURL(ctx context.Context, fileName string, duration time.Duration) (string, error)
	Download(ctx context.Context, fileName string) ([]byte, error)
	DeleteOldExports(ctx context.Context, duration time.Duration) error
}
