package driven

import (
	"context"
	"enfok_bd/src/domain/logs"
)

type LogRepository interface {
	Save(ctx context.Context, log *logs.ProcessingLog) error
	GetByImageID(ctx context.Context, imageUUID string) ([]logs.ProcessingLog, error)
}
