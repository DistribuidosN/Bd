package repository

import (
	"context"

	"enfok_bd/src/domain/metrics"
	"enfok_bd/src/domain/ports/driven"
	"enfok_bd/src/infrastructure/dto"
	"enfok_bd/src/utils"

	"github.com/jmoiron/sqlx"
)

type postgresMetricRepository struct {
	db *sqlx.DB
}

func NewPostgresMetricRepository(db *sqlx.DB) driven.MetricRepository {
	return &postgresMetricRepository{db: db}
}

func (r *postgresMetricRepository) Save(ctx context.Context, m *metrics.NodeMetrics) error {
	d := dto.MetricDTO{
		NodeID:          m.NodeID,
		ImageUUID:       m.ImageUUID,
		RamUsedMB:       m.RamUsedMB,
		RamTotalMB:      m.RamTotalMB,
		CpuPercent:      m.CpuPercent,
		WorkersBusy:     m.WorkersBusy,
		WorkersTotal:    m.WorkersTotal,
		QueueSize:       m.QueueSize,
		QueueCapacity:   m.QueueCapacity,
		TasksDone:       m.TasksDone,
		StealsPerformed: m.StealsPerformed,
		AvgLatencyMS:    m.AvgLatencyMS,
		P95LatencyMS:    m.P95LatencyMS,
		UptimeSeconds:   m.UptimeSeconds,
		StatusID:        utils.GetIDFromStatus(utils.NodeStatuses, m.Status),
		ReportedAt:      m.ReportedAt,
	}
	query := `INSERT INTO node_metrics (
		node_id, image_uuid, ram_used_mb, ram_total_mb, cpu_percent,
		workers_busy, workers_total, queue_size, queue_capacity,
		tasks_done, steals_performed, avg_latency_ms, p95_latency_ms,
		uptime_seconds, status_id, reported_at
	) VALUES (
		:node_id, :image_uuid, :ram_used_mb, :ram_total_mb, :cpu_percent,
		:workers_busy, :workers_total, :queue_size, :queue_capacity,
		:tasks_done, :steals_performed, :avg_latency_ms, :p95_latency_ms,
		:uptime_seconds, :status_id, :reported_at
	)`
	_, err := r.db.NamedExecContext(ctx, query, d)
	return err
}

func (r *postgresMetricRepository) GetByNodeID(ctx context.Context, nodeID int) ([]metrics.NodeMetrics, error) {
	var dtos []dto.MetricDTO
	if err := r.db.SelectContext(ctx, &dtos, `SELECT * FROM node_metrics WHERE node_id = $1 ORDER BY reported_at DESC LIMIT 100`, nodeID); err != nil {
		return nil, err
	}
	result := make([]metrics.NodeMetrics, 0, len(dtos))
	for _, d := range dtos {
		result = append(result, metrics.NodeMetrics{
			ID:              d.ID,
			NodeID:          d.NodeID,
			ImageUUID:       d.ImageUUID,
			RamUsedMB:       d.RamUsedMB,
			RamTotalMB:      d.RamTotalMB,
			CpuPercent:      d.CpuPercent,
			WorkersBusy:     d.WorkersBusy,
			WorkersTotal:    d.WorkersTotal,
			QueueSize:       d.QueueSize,
			QueueCapacity:   d.QueueCapacity,
			TasksDone:       d.TasksDone,
			StealsPerformed: d.StealsPerformed,
			AvgLatencyMS:    d.AvgLatencyMS,
			P95LatencyMS:    d.P95LatencyMS,
			UptimeSeconds:   d.UptimeSeconds,
			Status:          utils.GetStatusFromID(utils.NodeStatuses, d.StatusID),
			ReportedAt:      d.ReportedAt,
		})
	}
	return result, nil
}
