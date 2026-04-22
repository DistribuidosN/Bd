package driven

import "context"

type StorageRepository interface {
	Upload(ctx context.Context, fileName string, content []byte) (string, error)
	GetURL(ctx context.Context, fileName string) (string, error)
}
