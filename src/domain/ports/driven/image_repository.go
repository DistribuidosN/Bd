package driven

import (
	"context"
	"enfok_bd/src/domain/image"
	"time"
)

type ImageRepository interface {
	Save(ctx context.Context, img *image.Image) error
	GetByID(ctx context.Context, uuid string) (*image.Image, error)
	UpdateStatus(ctx context.Context, uuid string, status string) error
	UpdateResult(ctx context.Context, uuid string, resultPath string, nodeID string) error
	GetUserStatistics(ctx context.Context, userUUID string) (*UserStatistics, error)
	GetUserActivity(ctx context.Context, userUUID string) ([]UserActivity, error)
	ListByBatchPaginated(ctx context.Context, batchUUID string, limit, offset int) ([]image.Image, int, error)
}

type UserStatistics struct {
	TotalBatches  int
	TotalImages   int
	ImagesSuccess int
	ImagesFailed  int
}

type UserActivity struct {
	EventType  string
	RefUUID    string
	OccurredAt time.Time
}

type BatchWithCover struct {
	image.Batch
	CoverImageUUID  *string `json:"cover_image_uuid"`
	CoverResultPath *string `json:"cover_result_path"`
}

type BatchRepository interface {
	Save(ctx context.Context, batch *image.Batch) error
	GetByID(ctx context.Context, uuid string) (*image.Batch, error)
	UpdateStatus(ctx context.Context, uuid string, status string) error
	GetImagesByBatch(ctx context.Context, batchUUID string) ([]image.Image, error)
	SaveTransformations(ctx context.Context, batchUUID string, transformations []image.BatchTransformation) error
	GetTransformationsByBatch(ctx context.Context, batchUUID string) ([]image.BatchTransformation, error)
	ListByUserWithCover(ctx context.Context, userUUID string) ([]BatchWithCover, error)
}
