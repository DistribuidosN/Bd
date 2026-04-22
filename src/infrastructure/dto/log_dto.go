package dto

import "time"

// Level IDs para logs — deben coincidir con la tabla nivel_log en BD.
const (
	LogLevelInfo    = 1
	LogLevelWarning = 2
	LogLevelError   = 3
)

type LogDTO struct {
	ID        int       `db:"id"`
	NodeID    int       `db:"node_id"`
	ImageUUID string    `db:"image_uuid"`
	LevelID   int       `db:"level_id"`
	Message   string    `db:"message"`
	LogTime   time.Time `db:"log_time"`
}
