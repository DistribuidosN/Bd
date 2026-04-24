package service

import (
	"context"

	"enfok_bd/src/domain/logs"
	"enfok_bd/src/domain/ports/driven"
)

type LogService struct {
	repo driven.LogRepository
}

func NewLogService(r driven.LogRepository) *LogService {
	return &LogService{repo: r}
}

func (s *LogService) RegisterLog(ctx context.Context, l *logs.ProcessingLog) error {
	return s.repo.Save(ctx, l)
}

func (s *LogService) GetLogsByImage(ctx context.Context, imageUUID string) ([]logs.ProcessingLog, error) {
	return s.repo.GetByImageID(ctx, imageUUID)
}
