package dto

import "time"

// Status IDs para imágenes — deben coincidir con la tabla estatus_imagen en BD.
const (
	ImageStatusReceived   = 1
	ImageStatusProcessing = 2
	ImageStatusConverted  = 3
	ImageStatusFailed     = 4
)

// Status IDs para batches — deben coincidir con la tabla estatus_batch en BD.
const (
	BatchStatusPending    = 1
	BatchStatusProcessing = 2
	BatchStatusCompleted  = 3
	BatchStatusFailed     = 4
)

type ImageDTO struct {
	ImageUUID      string     `db:"image_uuid"`
	BatchUUID      string     `db:"batch_uuid"`
	OriginalName   string     `db:"original_name"`
	ResultPath     *string    `db:"result_path"`
	StatusID       int        `db:"status_id"`
	NodeID         *int       `db:"node_id"`
	ReceptionTime  time.Time  `db:"reception_time"`
	ConversionTime *time.Time `db:"conversion_time"`
}

type BatchDTO struct {
	BatchUUID   string    `db:"batch_uuid"`
	UserUUID    string    `db:"user_uuid"`
	RequestTime time.Time `db:"request_time"`
	StatusID    int       `db:"status_id"`
}
