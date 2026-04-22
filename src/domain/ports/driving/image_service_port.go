package driving

import (
	"context"

	"enfok_bd/src/domain/image"
)

// ImageServicePort define las operaciones que el API REST puede invocar
// sobre el dominio de imágenes. Es el contrato de entrada (driving).
type ImageServicePort interface {
	ProcessUpload(ctx context.Context, userUUID, fileName string, content []byte) (string, error)
	GetImage(ctx context.Context, uuid string) (*image.Image, error)
	UpdateStatus(ctx context.Context, uuid string, status string) error
}

// BatchServicePort define las operaciones que el API REST puede invocar
// sobre el dominio de batches. Es el contrato de entrada (driving).
type BatchServicePort interface {
	GetBatch(ctx context.Context, uuid string) (*image.Batch, error)
	UpdateStatus(ctx context.Context, uuid string, status string) error
}
