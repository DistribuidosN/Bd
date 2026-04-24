package service

import (
	"context"

	"enfok_bd/src/domain/metrics"
	"enfok_bd/src/domain/ports/driven"
)

type MetricsService struct {
	repo driven.MetricRepository
}

func NewMetricsService(r driven.MetricRepository) *MetricsService {
	return &MetricsService{repo: r}
}

func (s *MetricsService) SaveMetrics(ctx context.Context, m *metrics.NodeMetrics) error {
	return s.repo.Save(ctx, m)
}

func (s *MetricsService) GetMetricsByNode(ctx context.Context, nodeID string) ([]metrics.NodeMetrics, error) {
	return s.repo.GetByNodeID(ctx, nodeID)
}

func (s *MetricsService) GetMetricsByImage(ctx context.Context, imageUUID string) ([]metrics.NodeMetrics, error) {
	return s.repo.GetByImageUUID(ctx, imageUUID)
}
