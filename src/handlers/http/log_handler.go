package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"enfok_bd/src/domain/logs"
	"enfok_bd/src/domain/ports/driving"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	svc driving.LogServicePort
}

func NewLogHandler(s driving.LogServicePort) *LogHandler {
	return &LogHandler{svc: s}
}

type CreateLogRequest struct {
	NodeID    string  `json:"node_id" binding:"required"`
	ImageUUID *string `json:"image_uuid"`
	Level     string  `json:"level_name"` // Adjusted to match "level_name" from Java
	Message   string  `json:"message" binding:"required"`
	LogTime   string  `json:"log_time"`
}

func (h *LogHandler) CreateLog(c *gin.Context) {
	// Read raw body for debugging
	bodyBytes, _ := c.GetRawData()
	fmt.Printf("[DEBUG] Raw Logs JSON: %s\n", string(bodyBytes))

	// Put it back for binding
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req CreateLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[DEBUG] Logs Bind Error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid logs JSON schema",
			"details": err.Error(),
		})
		return
	}

	// Logic: If ImageUUID is empty string, treat as nil (NULL in DB)
	if req.ImageUUID != nil && *req.ImageUUID == "" {
		req.ImageUUID = nil
	}

	fmt.Printf("[DEBUG] Logs Parsed: %+v\n", req)

	// Convert string time to time.Time
	logTime := time.Now()
	if req.LogTime != "" {
		formats := []string{
			"2006-01-02T15:04:05.000-07:00",
			"2006-01-02T15:04:05.000Z07:00",
			"2006-01-02T15:04:05",
			time.RFC3339,
		}
		for _, f := range formats {
			if t, err := time.Parse(f, req.LogTime); err == nil {
				logTime = t
				break
			}
		}
	}

	l := logs.ProcessingLog{
		NodeID:    req.NodeID,
		ImageUUID: req.ImageUUID,
		Level:     req.Level,
		Message:   req.Message,
		LogTime:   &logTime,
	}

	if l.LogTime == nil || l.LogTime.IsZero() {
		now := time.Now()
		l.LogTime = &now
	}
	if err := h.svc.RegisterLog(c.Request.Context(), &l); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Log registered"})
}

func (h *LogHandler) GetLogsByImage(c *gin.Context) {
	result, err := h.svc.GetLogsByImage(c.Request.Context(), c.Param("image_uuid"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
