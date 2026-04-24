package driven

import (
	"context"
	"enfok_bd/src/domain/metrics"
)

type MetricRepository interface {
	Save(ctx context.Context, metric *metrics.NodeMetrics) error
	GetByNodeID(ctx context.Context, nodeID string) ([]metrics.NodeMetrics, error)
	GetByImageUUID(ctx context.Context, imageUUID string) ([]metrics.NodeMetrics, error)
}
