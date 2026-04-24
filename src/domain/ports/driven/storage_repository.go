package driven

import (
	"context"
	"time"
)

type StorageRepository interface {
	Upload(ctx context.Context, fileName string, content []byte) (string, error)
	GetURL(ctx context.Context, fileName string) (string, error)
	GetPresignedURL(ctx context.Context, fileName string, duration time.Duration) (string, error)
	Download(ctx context.Context, fileName string) ([]byte, error)
}
