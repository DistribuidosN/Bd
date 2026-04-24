package repository

import (
	"context"
	"fmt"
	"log"

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
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Buscar ID interno del nodo
	var internalNodeID int
	if err := tx.GetContext(ctx, &internalNodeID, "SELECT id FROM nodes WHERE node_id = $1", nodeID); err != nil {
		return fmt.Errorf("node %s not registered", nodeID)
	}

	// 2. Actualizar imagen
	queryImg := `UPDATE images SET result_path = $1, node_id = $2, status_id = 3, conversion_time = CURRENT_TIMESTAMP WHERE image_uuid = $3`
	if _, err := tx.ExecContext(ctx, queryImg, resultPath, internalNodeID, uuid); err != nil {
		return err
	}

	// 3. Obtener el batch_uuid de la imagen
	var batchUUID string
	if err := tx.GetContext(ctx, &batchUUID, "SELECT batch_uuid FROM images WHERE image_uuid = $1", uuid); err != nil {
		return err
	}

	// [IMPORTANTE]: Confirmamos la transacción aquí para que GetBatchProgress vea el cambio
	if err := tx.Commit(); err != nil {
		return err
	}

	// 4. Llamar a tu método de progreso
	progress, err := r.GetBatchProgress(ctx, batchUUID)
	if err == nil && progress.ProgressPercentage == 100 {
		// 5. Si está al 100%, actualizamos el estado del batch
		_, _ = r.db.ExecContext(ctx, "UPDATE batches SET status_id = 3 WHERE batch_uuid = $1", batchUUID)
	}

	return nil
}


func (r *postgresImageRepository) GetUserStatistics(ctx context.Context, userUUID string) (*driven.UserStatistics, error) {
	var stats driven.UserStatistics
	stats.TopTransformations = make([]driven.TransformationStat, 0)

	// Total Batches
	err := r.db.GetContext(ctx, &stats.TotalBatches, `SELECT COALESCE(COUNT(*), 0) FROM batches WHERE user_uuid = $1`, userUUID)
	if err != nil {
		return &driven.UserStatistics{TopTransformations: make([]driven.TransformationStat, 0)}, nil
	}

	// Total Images, Success, Failed
	queryCounts := `
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
	err = r.db.GetContext(ctx, &counts, queryCounts, userUUID)
	if err == nil {
		stats.TotalImages = counts.TotalImages
		stats.ImagesSuccess = counts.ImagesSuccess
		stats.ImagesFailed = counts.ImagesFailed
	}

	// Top Transformations
	queryTop := `
		SELECT tt.name, COUNT(*)::int as count
		FROM batch_transformations bt
		JOIN batches b ON bt.batch_uuid = b.batch_uuid
		JOIN transformation_types tt ON bt.type_id = tt.id
		WHERE b.user_uuid = $1
		GROUP BY tt.name
		ORDER BY count DESC
		LIMIT 5
	`
	err = r.db.SelectContext(ctx, &stats.TopTransformations, queryTop, userUUID)
	if err != nil {
		stats.TopTransformations = make([]driven.TransformationStat, 0)
	}

	log.Printf("[DEBUG GO] Datos extraídos de BD para Estadísticas de User %s: %+v\n", userUUID, stats)
	return &stats, nil
}

func (r *postgresImageRepository) GetUserActivity(ctx context.Context, userUUID string) ([]driven.UserActivity, error) {
	activities := make([]driven.UserActivity, 0)
	
	query := `
		-- Eventos de Creación de Lotes
		SELECT 
			'BATCH_CREATED' as event_type,
			b.batch_uuid as ref_uuid,
			'' as parent_uuid,
			'Lote creado con ' || (SELECT COUNT(*) FROM images WHERE batch_uuid = b.batch_uuid) || ' imágenes' as description,
			b.request_time as occurred_at
		FROM batches b
		WHERE b.user_uuid = $1

		UNION ALL

		-- Eventos de Imágenes Procesadas
		SELECT 
			'IMAGE_PROCESSED' as event_type,
			i.image_uuid as ref_uuid,
			i.batch_uuid as parent_uuid,
			'Imagen [' || i.original_name || '] procesada exitosamente por nodo ' || n.node_id as description,
			i.conversion_time as occurred_at
		FROM images i
		JOIN batches b ON i.batch_uuid = b.batch_uuid
		JOIN nodes n ON i.node_id = n.id
		WHERE b.user_uuid = $1 AND i.conversion_time IS NOT NULL

		ORDER BY occurred_at DESC
		LIMIT 50
	`
	err := r.db.SelectContext(ctx, &activities, query, userUUID)
	if err != nil {
		return make([]driven.UserActivity, 0), nil
	}
	log.Printf("[DEBUG GO] Datos extraídos de BD para Actividad de User %s: Total=%d, Datos=%+v\n", userUUID, len(activities), activities)
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
func (r *postgresImageRepository) GetBatchProgress(ctx context.Context, batchUUID string) (*driven.BatchProgress, error) {
	var stats struct {
		Total     int `db:"total"`
		Processed int `db:"processed"`
	}

	query := `
		SELECT 
			COUNT(*)::int as total,
			COUNT(CASE WHEN status_id IN (3, 4) THEN 1 END)::int as processed
		FROM images 
		WHERE batch_uuid = $1
	`
	if err := r.db.GetContext(ctx, &stats, query, batchUUID); err != nil {
		return nil, err
	}

	percentage := 0.0
	if stats.Total > 0 {
		percentage = (float64(stats.Processed) / float64(stats.Total)) * 100
	}

	return &driven.BatchProgress{
		BatchUUID:          batchUUID,
		TotalImages:        stats.Total,
		ProcessedImages:    stats.Processed,
		ProgressPercentage: percentage,
	}, nil
}
