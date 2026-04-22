package metrics

import "time"

type NodeMetrics struct {
	ID             int64     `json:"id"`
	NodeID         int       `json:"node_id"`
	ImageUUID      string    `json:"image_uuid"`
	RamUsedMB      float64   `json:"ram_used_mb"`
	RamTotalMB     float64   `json:"ram_total_mb"`
	CpuPercent     float64   `json:"cpu_percent"`
	WorkersBusy    int       `json:"workers_busy"`
	WorkersTotal   int       `json:"workers_total"`
	QueueSize      int       `json:"queue_size"`
	QueueCapacity  int       `json:"queue_capacity"`
	TasksDone      int       `json:"tasks_done"`
	StealsPerformed int      `json:"steals_performed"`
	AvgLatencyMS   *float64  `json:"avg_latency_ms,omitempty"`
	P95LatencyMS   *float64  `json:"p95_latency_ms,omitempty"`
	UptimeSeconds  int64     `json:"uptime_seconds"`
	Status         string    `json:"status"`
	ReportedAt     time.Time `json:"reported_at"`
}
