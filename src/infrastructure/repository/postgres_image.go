package repository

import (
	"context"

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
		NodeID:         img.NodeID,
		ReceptionTime:  img.ReceptionTime,
		ConversionTime: img.ConversionTime,
	}
	query := `INSERT INTO images (image_uuid, batch_uuid, original_name, result_path, status_id, node_id, reception_time, conversion_time) 
              VALUES (:image_uuid, :batch_uuid, :original_name, :result_path, :status_id, :node_id, :reception_time, :conversion_time)`
	_, err := r.db.NamedExecContext(ctx, query, d)
	return err
}

func (r *postgresImageRepository) GetByID(ctx context.Context, uuid string) (*image.Image, error) {
	var d dto.ImageDTO
	query := `SELECT image_uuid, batch_uuid, original_name, result_path, status_id, node_id, reception_time, conversion_time 
              FROM images WHERE image_uuid = $1`
	if err := r.db.GetContext(ctx, &d, query, uuid); err != nil {
		return nil, err
	}
	return &image.Image{
		ImageUUID:      d.ImageUUID,
		BatchUUID:      d.BatchUUID,
		OriginalName:   d.OriginalName,
		ResultPath:     d.ResultPath,
		Status:         utils.GetStatusFromID(utils.ImageStatuses, d.StatusID),
		NodeID:         d.NodeID,
		ReceptionTime:  d.ReceptionTime,
		ConversionTime: d.ConversionTime,
	}, nil
}

func (r *postgresImageRepository) UpdateStatus(ctx context.Context, uuid string, status string) error {
	statusID := utils.GetIDFromStatus(utils.ImageStatuses, status)
	_, err := r.db.ExecContext(ctx, `UPDATE images SET status_id = $1 WHERE image_uuid = $2`, statusID, uuid)
	return err
}

// --- Batch Repository ---

type postgresBatchRepository struct {
	db *sqlx.DB
}

func NewPostgresBatchRepository(db *sqlx.DB) driven.BatchRepository {
	return &postgresBatchRepository{db: db}
}

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
	}, nil
}

func (r *postgresBatchRepository) UpdateStatus(ctx context.Context, uuid string, status string) error {
	statusID := utils.GetIDFromStatus(utils.BatchStatuses, status)
	_, err := r.db.ExecContext(ctx, `UPDATE batches SET status_id = $1 WHERE batch_uuid = $2`, statusID, uuid)
	return err
}
