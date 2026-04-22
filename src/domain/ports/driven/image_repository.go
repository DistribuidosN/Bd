package driven

import (
	"context"
	"enfok_bd/src/domain/image"
)

type ImageRepository interface {
	Save(ctx context.Context, img *image.Image) error
	GetByID(ctx context.Context, uuid string) (*image.Image, error)
	UpdateStatus(ctx context.Context, uuid string, status string) error
}

type BatchRepository interface {
	Save(ctx context.Context, batch *image.Batch) error
	GetByID(ctx context.Context, uuid string) (*image.Batch, error)
	UpdateStatus(ctx context.Context, uuid string, status string) error
}
