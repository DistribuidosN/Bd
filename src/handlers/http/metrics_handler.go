package http

import (
	"net/http"
	"strconv"
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

func (h *MetricsHandler) CreateMetrics(c *gin.Context) {
	var m metrics.NodeMetrics
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if m.ReportedAt.IsZero() {
		m.ReportedAt = time.Now()
	}
	if err := h.svc.SaveMetrics(c.Request.Context(), &m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Metrics registered"})
}

func (h *MetricsHandler) GetMetricsByNode(c *gin.Context) {
	nodeID, _ := strconv.Atoi(c.Param("node_id"))
	result, err := h.svc.GetMetricsByNode(c.Request.Context(), nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
