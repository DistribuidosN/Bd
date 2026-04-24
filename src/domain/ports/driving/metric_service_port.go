package driving

import (
	"context"

	"enfok_bd/src/domain/metrics"
)

// MetricServicePort define las operaciones que el API REST puede invocar
// sobre el dominio de métricas. Es el contrato de entrada (driving).
type MetricServicePort interface {
	SaveMetrics(ctx context.Context, m *metrics.NodeMetrics) error
	GetMetricsByNode(ctx context.Context, nodeID string) ([]metrics.NodeMetrics, error)
}
