package driving

import (
	"context"

	"enfok_bd/src/domain/logs"
)

// LogServicePort define las operaciones que el API REST puede invocar
// sobre el dominio de logs. Es el contrato de entrada (driving).
type LogServicePort interface {
	RegisterLog(ctx context.Context, l *logs.ProcessingLog) error
	GetLogsByImage(ctx context.Context, imageUUID string) ([]logs.ProcessingLog, error)
}
