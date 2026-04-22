package service

import (
	"context"

	"enfok_bd/src/domain/image"
	"enfok_bd/src/domain/ports/driven"
)

type BatchService struct {
	repo driven.BatchRepository
}

func NewBatchService(r driven.BatchRepository) *BatchService {
	return &BatchService{repo: r}
}

func (s *BatchService) GetBatch(ctx context.Context, uuid string) (*image.Batch, error) {
	return s.repo.GetByID(ctx, uuid)
}

func (s *BatchService) UpdateStatus(ctx context.Context, uuid string, status string) error {
	return s.repo.UpdateStatus(ctx, uuid, status)
}
