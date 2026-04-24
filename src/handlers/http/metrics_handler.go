package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"enfok_bd/src/domain/metrics"
	"enfok_bd/src/domain/ports/driving"

	"github.com/gin-gonic/gin"
)

type MetricsHandler struct {
	svc driving.MetricServicePort
}

func NewMetricsHandler(s driving.MetricServicePort) *MetricsHandler {
	return &MetricsHandler{svc: s}
}

type CreateMetricRequest struct {
	NodeID          string   `json:"node_id" binding:"required"` // Strictly string
	ImageUUID       *string  `json:"image_uuid"`
	RamUsedMB       float64  `json:"ram_used_mb"`
	RamTotalMB      float64  `json:"ram_total_mb"`
	CpuPercent      float64  `json:"cpu_percent"`
	WorkersBusy     int      `json:"workers_busy"`
	WorkersTotal    int      `json:"workers_total"`
	QueueSize       int      `json:"queue_size"`
	QueueCapacity   int      `json:"queue_capacity"`
	TasksDone       int      `json:"tasks_done"`
	StealsPerformed int      `json:"steals_performed"`
	AvgLatencyMS    *float64 `json:"avg_latency_ms"`
	P95LatencyMS    *float64 `json:"p95_latency_ms"`
	UptimeSeconds   int64    `json:"uptime_seconds"`
	Status          string   `json:"status"`
	ReportedAt      string   `json:"reported_at"`
}

func (h *MetricsHandler) CreateMetrics(c *gin.Context) {
	// Read raw body for debugging
	bodyBytes, _ := c.GetRawData()
	fmt.Printf("[DEBUG] Raw Metrics JSON: %s\n", string(bodyBytes))

	// Put it back for binding
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req CreateMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[DEBUG] Metrics Bind Error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid metrics JSON schema",
			"details": err.Error(),
		})
		return
	}

	// Logic: If ImageUUID is empty string, treat as nil (NULL in DB)
	if req.ImageUUID != nil && *req.ImageUUID == "" {
		req.ImageUUID = nil
	}

	fmt.Printf("[DEBUG] Metrics Parsed: %+v\n", req)

	// Convert string time to time.Time
	reportedAt := time.Now()
	if req.ReportedAt != "" {
		// Java formats can be: "2026-04-24T05:08:33.978+00:00" or "2026-04-24T05:08:33"
		formats := []string{
			"2006-01-02T15:04:05.000-07:00",
			"2006-01-02T15:04:05.000Z07:00",
			"2006-01-02T15:04:05",
			time.RFC3339,
		}
		for _, f := range formats {
			if t, err := time.Parse(f, req.ReportedAt); err == nil {
				reportedAt = t
				break
			}
		}
	}

	m := metrics.NodeMetrics{
		NodeID:          req.NodeID,
		ImageUUID:       req.ImageUUID,
		RamUsedMB:       req.RamUsedMB,
		RamTotalMB:      req.RamTotalMB,
		CpuPercent:      req.CpuPercent,
		WorkersBusy:     req.WorkersBusy,
		WorkersTotal:    req.WorkersTotal,
		QueueSize:       req.QueueSize,
		QueueCapacity:   req.QueueCapacity,
		TasksDone:       req.TasksDone,
		StealsPerformed: req.StealsPerformed,
		AvgLatencyMS:    req.AvgLatencyMS,
		P95LatencyMS:    req.P95LatencyMS,
		UptimeSeconds:   req.UptimeSeconds,
		Status:          req.Status,
		ReportedAt:      reportedAt,
	}

	if err := h.svc.SaveMetrics(c.Request.Context(), &m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Metrics registered"})
}

func (h *MetricsHandler) GetMetricsByNode(c *gin.Context) {
	nodeID := c.Param("node_id") // Use as string
	result, err := h.svc.GetMetricsByNode(c.Request.Context(), nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.PureJSON(http.StatusOK, result)
}

func (h *MetricsHandler) GetMetricsByImage(c *gin.Context) {
	imageUUID := c.Param("image_uuid")
	result, err := h.svc.GetMetricsByImage(c.Request.Context(), imageUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Senior requirement: ensure [] in JSON, which repository already does with make()
	c.PureJSON(http.StatusOK, result)
}
