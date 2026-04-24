package repository

import (
	"context"
	"fmt"

	"enfok_bd/src/domain/image"
	"enfok_bd/src/domain/ports/driven"
	"enfok_bd/src/infrastructure/dto"
	"enfok_bd/src/utils"

	"github.com/jmoiron/sqlx"
)

// --- Image Repository ---

type postgresImageRepository struct {
	db *sqlx.DB
}

func NewPostgresImageRepository(db *sqlx.DB) driven.ImageRepository {
	return &postgresImageRepository{db: db}
}

func (r *postgresImageRepository) Save(ctx context.Context, img *image.Image) error {
	d := dto.ImageDTO{
		ImageUUID:      img.ImageUUID,
		BatchUUID:      img.BatchUUID,
		OriginalName:   img.OriginalName,
		ResultPath:     img.ResultPath,
		StatusID:       utils.GetIDFromStatus(utils.ImageStatuses, img.Status),
		// NodeID se maneja como NULL en el registro inicial
		ReceptionTime:  img.ReceptionTime,
		ConversionTime: img.ConversionTime,
	}
	query := `INSERT INTO images (image_uuid, batch_uuid, original_name, result_path, status_id, reception_time, conversion_time) 
              VALUES (:image_uuid, :batch_uuid, :original_name, :result_path, :status_id, :reception_time, :conversion_time)`
	_, err := r.db.NamedExecContext(ctx, query, d)
	return err
}

func (r *postgresImageRepository) GetByID(ctx context.Context, uuid string) (*image.Image, error) {
	var d dto.ImageDTO
	query := `SELECT i.image_uuid, i.batch_uuid, i.original_name, i.result_path, i.status_id, i.node_id, n.node_id as node_name, i.reception_time, i.conversion_time 
              FROM images i 
              LEFT JOIN nodes n ON i.node_id = n.id
              WHERE i.image_uuid = $1`
	if err := r.db.GetContext(ctx, &d, query, uuid); err != nil {
		return nil, err
	}

	nodeName := ""
	if d.NodeName != nil {
		nodeName = *d.NodeName
	}

	return &image.Image{
		ImageUUID:      d.ImageUUID,
		BatchUUID:      d.BatchUUID,
		OriginalName:   d.OriginalName,
		ResultPath:     d.ResultPath,
		Status:         utils.GetStatusFromID(utils.ImageStatuses, d.StatusID),
		StatusID:       d.StatusID,
		NodeID:         nodeName,
		ReceptionTime:  d.ReceptionTime,
		ConversionTime: d.ConversionTime,
	}, nil
}

func (r *postgresImageRepository) UpdateStatus(ctx context.Context, uuid string, status string) error {
	statusID := utils.GetIDFromStatus(utils.ImageStatuses, status)
	_, err := r.db.ExecContext(ctx, `UPDATE images SET status_id = $1 WHERE image_uuid = $2`, statusID, uuid)
	return err
}

func (r *postgresImageRepository) UpdateResult(ctx context.Context, uuid string, resultPath string, nodeID string) error {
	var internalID int
	err := r.db.GetContext(ctx, &internalID, "SELECT id FROM nodes WHERE node_id = $1", nodeID)
	if err != nil {
		return fmt.Errorf("node %s not registered", nodeID)
	}

	query := `UPDATE images SET result_path = $1, status_id = 3, node_id = $2, conversion_time = CURRENT_TIMESTAMP WHERE image_uuid = $3`
	_, err = r.db.ExecContext(ctx, query, resultPath, internalID, uuid)
	return err
}

func (r *postgresImageRepository) GetUserStatistics(ctx context.Context, userUUID string) (*driven.UserStatistics, error) {
	var stats driven.UserStatistics

	// Total Batches con COALESCE para seguridad
	err := r.db.GetContext(ctx, &stats.TotalBatches, `SELECT COALESCE(COUNT(*), 0) FROM batches WHERE user_uuid = $1`, userUUID)
	if err != nil {
		return &driven.UserStatistics{}, nil // Devolvemos ceros en lugar de error 500
	}

	// Total Images, Success, Failed con COALESCE
	query := `
		SELECT 
			COALESCE(COUNT(*), 0) as total_images,
			COALESCE(COUNT(CASE WHEN status_id = 3 THEN 1 END), 0) as images_success,
			COALESCE(COUNT(CASE WHEN status_id = 4 THEN 1 END), 0) as images_failed
		FROM images i
		JOIN batches b ON i.batch_uuid = b.batch_uuid
		WHERE b.user_uuid = $1
	`
	var counts struct {
		TotalImages   int `db:"total_images"`
		ImagesSuccess int `db:"images_success"`
		ImagesFailed  int `db:"images_failed"`
	}
	err = r.db.GetContext(ctx, &counts, query, userUUID)
	if err != nil {
		return &stats, nil // Devolvemos lo que tengamos (probablemente ceros)
	}

	stats.TotalImages = counts.TotalImages
	stats.ImagesSuccess = counts.ImagesSuccess
	stats.ImagesFailed = counts.ImagesFailed

	return &stats, nil
}

func (r *postgresImageRepository) GetUserActivity(ctx context.Context, userUUID string) ([]driven.UserActivity, error) {
	// Inicializamos con make para que en JSON sea [] y no null
	activities := make([]driven.UserActivity, 0)
	
	query := `
		SELECT 'BATCH_CREATED' as event_type, batch_uuid as ref_uuid, COALESCE(request_time, CURRENT_TIMESTAMP) as occurred_at
		FROM batches
		WHERE user_uuid = $1
		UNION ALL
		SELECT 'IMAGE_PROCESSED' as event_type, i.image_uuid as ref_uuid, COALESCE(i.conversion_time, CURRENT_TIMESTAMP) as occurred_at
		FROM images i
		JOIN batches b ON i.batch_uuid = b.batch_uuid
		WHERE b.user_uuid = $1 AND i.conversion_time IS NOT NULL
		ORDER BY occurred_at DESC
		LIMIT 20
	`
	err := r.db.SelectContext(ctx, &activities, query, userUUID)
	if err != nil {
		return make([]driven.UserActivity, 0), nil // Siempre devolvemos lista vacía, nunca error 500
	}
	return activities, nil
}

func (r *postgresImageRepository) ListByBatchPaginated(ctx context.Context, batchUUID string, limit, offset int) ([]image.Image, int, error) {
	// Inicialización Senior: aseguramos [] en el JSON
	images := make([]image.Image, 0)
	var total int
	
	err := r.db.GetContext(ctx, &total, `SELECT COALESCE(COUNT(*), 0) FROM images WHERE batch_uuid = $1`, batchUUID)
	if err != nil {
		return images, 0, nil
	}

	var dtos []dto.ImageDTO
	query := `SELECT i.image_uuid, i.batch_uuid, i.original_name, i.result_path, i.status_id, i.node_id, n.node_id as node_name, i.reception_time, i.conversion_time 
              FROM images i 
              LEFT JOIN nodes n ON i.node_id = n.id
              WHERE i.batch_uuid = $1 ORDER BY i.reception_time ASC LIMIT $2 OFFSET $3`
	
	if err := r.db.SelectContext(ctx, &dtos, query, batchUUID, limit, offset); err != nil {
		return images, total, nil
	}

	images = make([]image.Image, 0, len(dtos))
	for _, d := range dtos {
		nodeName := ""
		if d.NodeName != nil {
			nodeName = *d.NodeName
		}

		images = append(images, image.Image{
			ImageUUID:      d.ImageUUID,
			BatchUUID:      d.BatchUUID,
			OriginalName:   d.OriginalName,
			ResultPath:     d.ResultPath,
			Status:         utils.GetStatusFromID(utils.ImageStatuses, d.StatusID),
			StatusID:       d.StatusID,
			NodeID:         nodeName,
			ReceptionTime:  d.ReceptionTime,
			ConversionTime: d.ConversionTime,
		})
	}
	return images, total, nil
}
