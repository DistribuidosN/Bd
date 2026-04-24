package driving

import (
	"context"
	"time"

	"enfok_bd/src/domain/image"
)

// ImageServicePort define las operaciones que el API REST puede invocar
// sobre el dominio de imágenes. Es el contrato de entrada (driving).
type ImageServicePort interface {
	ProcessUpload(ctx context.Context, userUUID, fileName string, content []byte) (string, error)
	UploadBatch(ctx context.Context, userUUID string, files []BatchFile) (string, []string, error)
	GetImage(ctx context.Context, uuid string) (*image.Image, error)
	UpdateStatus(ctx context.Context, uuid string, status string) error
	UpdateResult(ctx context.Context, uuid string, content []byte, nodeID string) (string, error)
	RegisterImage(ctx context.Context, userUUID, batchUUID, imageUUID, fileName string) error
	RegisterBatch(ctx context.Context, userUUID, batchUUID string) error
	GetBatchImagesPaginated(ctx context.Context, batchUUID string, page, limit int) (*PaginatedImages, error)
	GetUserStatistics(ctx context.Context, userUUID string) (*UserStatistics, error)
	GetUserActivity(ctx context.Context, userUUID string) ([]UserActivity, error)
	DownloadImage(ctx context.Context, uuid string) ([]byte, string, error)
}

// BatchFile representa un archivo recibido en una subida de batch.
type BatchFile struct {
	FileName string
	Content  []byte
}

type PaginatedImages struct {
	Images      []image.Image `json:"images"`
	CurrentPage int           `json:"current_page" json:"currentPage"`
	Limit       int           `json:"limit"`
	TotalCount  int           `json:"total_count" json:"totalCount"`
	HasMore     bool          `json:"has_more" json:"hasMore"`
}

type UserStatistics struct {
	TotalBatches  int `json:"total_batches"`
	TotalImages   int `json:"total_images"`
	ImagesSuccess int `json:"images_success"`
	ImagesFailed  int `json:"images_failed"`
}

type UserActivity struct {
	EventType  string    `json:"event_type"`
	RefUUID    string    `json:"ref_uuid"`
	OccurredAt time.Time `json:"occurred_at"`
}

type BatchWithCover struct {
	Batch           image.Batch `json:"batch"`
	CoverImageURL   string      `json:"cover_image_url"`
	CoverImageUUID  string      `json:"cover_image_uuid"`
}

// BatchServicePort define las operaciones que el API REST puede invocar
// sobre el dominio de batches. Es el contrato de entrada (driving).
type BatchServicePort interface {
	GetBatch(ctx context.Context, uuid string) (*image.Batch, error)
	UpdateStatus(ctx context.Context, uuid string, status string) error
	GetBatchWithImages(ctx context.Context, uuid string) (*image.Batch, []image.Image, error)
	SaveTransformations(ctx context.Context, batchUUID string, transformations []image.BatchTransformation) error
	ListUserBatches(ctx context.Context, userUUID string) ([]BatchWithCover, error)
}
