package repository

import (
	"context"
	"fmt"
	"log"

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
	var internalID int
	fmt.Printf("[DEBUG] Metric Repo: Looking up internal ID for node name: %s\n", m.NodeID)
	err := r.db.GetContext(ctx, &internalID, "SELECT id FROM nodes WHERE node_id = $1", m.NodeID)
	if err != nil {
		fmt.Printf("[DEBUG] Metric Repo Error: Node name %s not found in 'nodes' table: %v\n", m.NodeID, err)
		return fmt.Errorf("node %s not registered", m.NodeID)
	}

	d := dto.MetricDTO{
		NodeID:          internalID,
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
	fmt.Printf("[DEBUG] Metric Repo: Saving with ImageUUID: %v\n", d.ImageUUID)
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
	_, err = r.db.NamedExecContext(ctx, query, d)
	return err
}

func (r *postgresMetricRepository) GetByNodeID(ctx context.Context, nodeName string) ([]metrics.NodeMetrics, error) {
	// Inicializamos para que nunca sea null en JSON
	result := make([]metrics.NodeMetrics, 0)
	
	var internalID int
	err := r.db.GetContext(ctx, &internalID, "SELECT id FROM nodes WHERE node_id = $1", nodeName)
	if err != nil {
		fmt.Printf("[DEBUG] Metric Repo: Node '%s' not found for metrics retrieval\n", nodeName)
		return make([]metrics.NodeMetrics, 0), nil // Siempre []
	}

	var dtos []dto.MetricDTO
	query := `SELECT id, node_id, image_uuid, ram_used_mb, ram_total_mb, cpu_percent, 
	          workers_busy, workers_total, queue_size, queue_capacity, tasks_done, 
	          steals_performed, avg_latency_ms, p95_latency_ms, uptime_seconds, 
	          status_id, COALESCE(reported_at, CURRENT_TIMESTAMP) as reported_at 
	          FROM node_metrics WHERE node_id = $1 ORDER BY reported_at DESC LIMIT 100`
	
	if err := r.db.SelectContext(ctx, &dtos, query, internalID); err != nil {
		return make([]metrics.NodeMetrics, 0), nil
	}

	for _, d := range dtos {
		result = append(result, metrics.NodeMetrics{
			ID:              d.ID,
			NodeID:          nodeName,
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
	log.Printf("[DEBUG GO] Datos extraídos de BD para Métricas del Nodo %s: Total=%d, Datos=%+v\n", nodeName, len(result), result)
	return result, nil
}
func (r *postgresMetricRepository) GetByImageUUID(ctx context.Context, imageUUID string) ([]metrics.NodeMetrics, error) {
	result := make([]metrics.NodeMetrics, 0)

	var dtos []dto.MetricDTO
	query := `SELECT m.id, n.node_id as node_name, m.image_uuid, m.ram_used_mb, m.ram_total_mb, m.cpu_percent, 
	          m.workers_busy, m.workers_total, m.queue_size, m.queue_capacity, m.tasks_done, 
	          m.steals_performed, m.avg_latency_ms, m.p95_latency_ms, m.uptime_seconds, 
	          m.status_id, COALESCE(m.reported_at, CURRENT_TIMESTAMP) as reported_at 
	          FROM node_metrics m
	          JOIN nodes n ON m.node_id = n.id
	          WHERE m.image_uuid = $1 ORDER BY m.reported_at DESC`

	if err := r.db.SelectContext(ctx, &dtos, query, imageUUID); err != nil {
		return make([]metrics.NodeMetrics, 0), nil
	}

	for _, d := range dtos {
		nodeName := ""
		if d.NodeName != nil {
			nodeName = *d.NodeName
		}
		result = append(result, metrics.NodeMetrics{
			ID:              d.ID,
			NodeID:          nodeName,
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
	log.Printf("[DEBUG GO] Datos extraídos de BD para Métricas de Imagen %s: Total=%d, Datos=%+v\n", imageUUID, len(result), result)
	return result, nil
}
