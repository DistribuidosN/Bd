package repository

import (
	"context"
	"fmt"
	"log"
	"strconv"

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
	var internalNodeID int
	fmt.Printf("[DEBUG] Log Repo: Looking up node_id: %s\n", l.NodeID)
	err := r.db.GetContext(ctx, &internalNodeID, "SELECT id FROM nodes WHERE node_id = $1", l.NodeID)
	if err != nil {
		fmt.Printf("[DEBUG] Log Repo Error: Node %s not found: %v\n", l.NodeID, err)
		return err
	}
	fmt.Printf("[DEBUG] Log Repo: Resolved internal ID: %d\n", internalNodeID)

	d := dto.LogDTO{
		NodeID:    internalNodeID,
		ImageUUID: l.ImageUUID,
		LevelID:   utils.GetIDFromStatus(utils.LogLevels, l.Level),
		Message:   l.Message,
		LogTime:   l.LogTime,
	}
	query := `INSERT INTO processing_logs (node_id, image_uuid, level_id, message, log_time) 
              VALUES (:node_id, :image_uuid, :level_id, :message, :log_time)`
	_, err = r.db.NamedExecContext(ctx, query, d)
	if err != nil {
		fmt.Printf("[DEBUG] Log Repo INSERT Error: %v\n", err)
	}
	return err
}

func (r *postgresLogRepository) GetByImageID(ctx context.Context, uuid string) ([]logs.ProcessingLog, error) {
	// Inicializamos para evitar null en JSON
	result := make([]logs.ProcessingLog, 0)

	var dtos []dto.LogDTO
	query := `SELECT id, node_id, image_uuid, level_id, message, COALESCE(log_time, CURRENT_TIMESTAMP) as log_time 
              FROM processing_logs WHERE image_uuid = $1 ORDER BY log_time ASC`
	if err := r.db.SelectContext(ctx, &dtos, query, uuid); err != nil {
		return result, nil // Silenciamos error para evitar 500
	}
	// ...
	for _, d := range dtos {
		result = append(result, logs.ProcessingLog{
			ID:        d.ID,
			NodeID:    strconv.Itoa(d.NodeID),
			ImageUUID: d.ImageUUID,
			Level:     utils.GetStatusFromID(utils.LogLevels, d.LevelID),
			Message:   d.Message,
			LogTime:   d.LogTime,
		})
	}
	log.Printf("[DEBUG GO] Datos extraídos de BD para Logs de Imagen %s: Total=%d, Datos=%+v\n", uuid, len(result), result)
	return result, nil
}
