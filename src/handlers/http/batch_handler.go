package http

import (
	"net/http"

	"enfok_bd/src/domain/ports/driving"

	"github.com/gin-gonic/gin"
)

type BatchHandler struct {
	svc driving.BatchServicePort
}

func NewBatchHandler(s driving.BatchServicePort) *BatchHandler {
	return &BatchHandler{svc: s}
}

func (h *BatchHandler) GetBatch(c *gin.Context) {
	b, err := h.svc.GetBatch(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Batch not found"})
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *BatchHandler) UpdateStatus(c *gin.Context) {
	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), c.Param("id"), body.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Batch status updated"})
}
