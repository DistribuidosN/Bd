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
	GetBatchProgress(ctx context.Context, batchUUID string) (*BatchProgress, error)
}

type BatchProgress struct {
	BatchUUID          string  `json:"batch_uuid"`
	TotalImages        int     `json:"total_images"`
	ProcessedImages    int     `json:"processed_images"`
	ProgressPercentage float64 `json:"progress_percentage"`
}

type TransformationStat struct {
	Name  string `json:"name" db:"name"`
	Count int    `json:"count" db:"count"`
}

type UserStatistics struct {
	TotalBatches       int                  `json:"total_batches"`
	TotalImages        int                  `json:"total_images"`
	ImagesSuccess      int                  `json:"images_success"`
	ImagesFailed       int                  `json:"images_failed"`
	TopTransformations []TransformationStat `json:"top_transformations"`
}

type UserActivity struct {
	EventType   string    `json:"event_type" db:"event_type"`
	RefUUID     string    `json:"ref_uuid" db:"ref_uuid"`
	ParentUUID  string    `json:"parent_uuid,omitempty" db:"parent_uuid"`
	Description string    `json:"description" db:"description"`
	OccurredAt  time.Time `json:"occurred_at" db:"occurred_at"`
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
