package http

import (
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

func (h *LogHandler) CreateLog(c *gin.Context) {
	var l logs.ProcessingLog
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if l.LogTime.IsZero() {
		l.LogTime = time.Now()
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
