package dto

import "time"

// Status IDs para métricas de nodo (reutiliza NodeStatus de node_dto.go).
// Los status de métricas representan el estado del nodo en ese momento.
type MetricDTO struct {
	ID              int64    `db:"id"`
	NodeID          int      `db:"node_id"`
	ImageUUID       *string  `db:"image_uuid"`
	RamUsedMB       float64  `db:"ram_used_mb"`
	RamTotalMB      float64  `db:"ram_total_mb"`
	CpuPercent      float64  `db:"cpu_percent"`
	WorkersBusy     int      `db:"workers_busy"`
	WorkersTotal    int      `db:"workers_total"`
	QueueSize       int      `db:"queue_size"`
	QueueCapacity   int      `db:"queue_capacity"`
	TasksDone       int      `db:"tasks_done"`
	StealsPerformed int      `db:"steals_performed"`
	AvgLatencyMS    *float64 `db:"avg_latency_ms"`
	P95LatencyMS    *float64 `db:"p95_latency_ms"`
	UptimeSeconds   int64    `db:"uptime_seconds"`
	StatusID        int      `db:"status_id"`
	ReportedAt      time.Time `db:"reported_at"`
}
