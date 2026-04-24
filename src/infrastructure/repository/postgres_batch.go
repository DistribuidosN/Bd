package repository

import (
	"context"

	"enfok_bd/src/domain/image"
	"enfok_bd/src/domain/ports/driven"
	"enfok_bd/src/infrastructure/dto"
	"enfok_bd/src/utils"

	"github.com/jmoiron/sqlx"
)

type postgresBatchRepository struct {
	db *sqlx.DB
}

func NewPostgresBatchRepository(db *sqlx.DB) driven.BatchRepository {
	// Nota: Implementación extendida con ListByUserWithCover
	return &postgresBatchRepository{db: db}
}

// ... Implementaremos solo los nuevos métodos o todos si es necesario ...
// Para este ejercicio, implementaré el método solicitado.

func (r *postgresBatchRepository) ListByUserWithCover(ctx context.Context, userUUID string) ([]driven.BatchWithCover, error) {
	query := `
		SELECT DISTINCT ON (b.batch_uuid)
			b.batch_uuid, b.user_uuid, b.status_id, b.request_time,
			i.image_uuid as cover_image_uuid, i.result_path as cover_result_path
		FROM batches b
		LEFT JOIN images i ON b.batch_uuid = i.batch_uuid
		WHERE b.user_uuid = $1
		ORDER BY b.batch_uuid, RANDOM()
	`
	
	var rows []struct {
		dto.BatchDTO
		CoverImageUUID  *string `db:"cover_image_uuid"`
		CoverResultPath *string `db:"cover_result_path"`
	}

	if err := r.db.SelectContext(ctx, &rows, query, userUUID); err != nil {
		return make([]driven.BatchWithCover, 0), nil
	}

	result := make([]driven.BatchWithCover, 0)
	for _, row := range rows {
		result = append(result, driven.BatchWithCover{
			Batch: image.Batch{
				BatchUUID:   row.BatchUUID,
				UserUUID:    row.UserUUID,
				RequestTime: row.RequestTime,
				Status:      utils.GetStatusFromID(utils.BatchStatuses, row.StatusID),
				StatusID:    row.StatusID,
			},
			CoverImageUUID:  row.CoverImageUUID,
			CoverResultPath: row.CoverResultPath,
		})
	}
	return result, nil
}

// Para que compile como driven.BatchRepository, necesita el resto de métodos.
// Los delegaremos o los moveremos aquí. Por simplicidad, los copiaré del original.

func (r *postgresBatchRepository) Save(ctx context.Context, b *image.Batch) error {
	d := dto.BatchDTO{
		BatchUUID:   b.BatchUUID,
		UserUUID:    b.UserUUID,
		RequestTime: b.RequestTime,
		StatusID:    utils.GetIDFromStatus(utils.BatchStatuses, b.Status),
	}
	query := `INSERT INTO batches (batch_uuid, user_uuid, request_time, status_id) 
              VALUES (:batch_uuid, :user_uuid, :request_time, :status_id)`
	_, err := r.db.NamedExecContext(ctx, query, d)
	return err
}

func (r *postgresBatchRepository) GetByID(ctx context.Context, uuid string) (*image.Batch, error) {
	var d dto.BatchDTO
	if err := r.db.GetContext(ctx, &d, `SELECT batch_uuid, user_uuid, request_time, status_id FROM batches WHERE batch_uuid = $1`, uuid); err != nil {
		return nil, err
	}
	return &image.Batch{
		BatchUUID:   d.BatchUUID,
		UserUUID:    d.UserUUID,
		RequestTime: d.RequestTime,
		Status:      utils.GetStatusFromID(utils.BatchStatuses, d.StatusID),
		StatusID:    d.StatusID,
	}, nil
}

func (r *postgresBatchRepository) UpdateStatus(ctx context.Context, uuid string, status string) error {
	statusID := utils.GetIDFromStatus(utils.BatchStatuses, status)
	_, err := r.db.ExecContext(ctx, `UPDATE batches SET status_id = $1 WHERE batch_uuid = $2`, statusID, uuid)
	return err
}

func (r *postgresBatchRepository) GetImagesByBatch(ctx context.Context, batchUUID string) ([]image.Image, error) {
	var dtos []dto.ImageDTO
	query := `SELECT i.image_uuid, i.batch_uuid, i.original_name, i.result_path, i.status_id, i.node_id, n.node_id as node_name, i.reception_time, i.conversion_time
              FROM images i
              LEFT JOIN nodes n ON i.node_id = n.id
              WHERE i.batch_uuid = $1 ORDER BY i.reception_time ASC`
	if err := r.db.SelectContext(ctx, &dtos, query, batchUUID); err != nil {
		return make([]image.Image, 0), nil
	}
	res := make([]image.Image, 0)
	for _, d := range dtos {
		nodeName := ""
		if d.NodeName != nil {
			nodeName = *d.NodeName
		}

		res = append(res, image.Image{
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
	return res, nil
}

func (r *postgresBatchRepository) SaveTransformations(ctx context.Context, batchUUID string, transformations []image.BatchTransformation) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO batch_transformations (batch_uuid, type_id, params, execution_order) 
              VALUES (:batch_uuid, :type_id, :params, :execution_order)`

	for _, t := range transformations {
		d := dto.BatchTransformationDTO{
			BatchUUID:      batchUUID,
			TypeID:         t.TypeID,
			Params:         t.Params,
			ExecutionOrder: t.ExecutionOrder,
		}
		if _, err := tx.NamedExecContext(ctx, query, d); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *postgresBatchRepository) GetTransformationsByBatch(ctx context.Context, batchUUID string) ([]image.BatchTransformation, error) {
	var dtos []dto.BatchTransformationDTO
	query := `SELECT batch_uuid, type_id, params, execution_order FROM batch_transformations WHERE batch_uuid = $1 ORDER BY execution_order ASC`
	if err := r.db.SelectContext(ctx, &dtos, query, batchUUID); err != nil {
		return make([]image.BatchTransformation, 0), nil
	}
	res := make([]image.BatchTransformation, 0)
	for _, d := range dtos {
		res = append(res, image.BatchTransformation{
			BatchUUID:      d.BatchUUID,
			TypeID:         d.TypeID,
			Params:         d.Params,
			ExecutionOrder: d.ExecutionOrder,
		})
	}
	return res, nil
}
