package repository

import (
	"context"

	"enfok_bd/src/domain/logs"
	"enfok_bd/src/domain/ports/driven"
	"enfok_bd/src/infrastructure/dto"
	"enfok_bd/src/utils"

	"github.com/jmoiron/sqlx"
)

type postgresLogRepository struct {
	db *sqlx.DB
}

func NewPostgresLogRepository(db *sqlx.DB) driven.LogRepository {
	return &postgresLogRepository{db: db}
}

func (r *postgresLogRepository) Save(ctx context.Context, l *logs.ProcessingLog) error {
	d := dto.LogDTO{
		NodeID:    l.NodeID,
		ImageUUID: l.ImageUUID,
		LevelID:   utils.GetIDFromStatus(utils.LogLevels, l.Level),
		Message:   l.Message,
		LogTime:   l.LogTime,
	}
	query := `INSERT INTO processing_logs (node_id, image_uuid, level_id, message, log_time) 
              VALUES (:node_id, :image_uuid, :level_id, :message, :log_time)`
	_, err := r.db.NamedExecContext(ctx, query, d)
	return err
}

func (r *postgresLogRepository) GetByImageID(ctx context.Context, uuid string) ([]logs.ProcessingLog, error) {
	var dtos []dto.LogDTO
	query := `SELECT id, node_id, image_uuid, level_id, message, log_time 
              FROM processing_logs WHERE image_uuid = $1 ORDER BY log_time ASC`
	if err := r.db.SelectContext(ctx, &dtos, query, uuid); err != nil {
		return nil, err
	}
	result := make([]logs.ProcessingLog, 0, len(dtos))
	for _, d := range dtos {
		result = append(result, logs.ProcessingLog{
			ID:        d.ID,
			NodeID:    d.NodeID,
			ImageUUID: d.ImageUUID,
			Level:     utils.GetStatusFromID(utils.LogLevels, d.LevelID),
			Message:   d.Message,
			LogTime:   d.LogTime,
		})
	}
	return result, nil
}
