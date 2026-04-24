package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"enfok_bd/src/domain/image"
	"enfok_bd/src/domain/ports/driving"
	"enfok_bd/src/utils"

	"github.com/gin-gonic/gin"
)

type BatchHandler struct {
	svc driving.BatchServicePort
}

func NewBatchHandler(s driving.BatchServicePort) *BatchHandler {
	return &BatchHandler{svc: s}
}

func (h *BatchHandler) ListBatches(c *gin.Context) {
	userUUID := c.Query("user_uuid")
	if userUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_uuid is required"})
		return
	}
	batches, err := h.svc.ListUserBatches(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, batches)
}

func (h *BatchHandler) GetBatch(c *gin.Context) {
	b, err := h.svc.GetBatch(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Batch not found"})
		return
	}
	c.JSON(http.StatusOK, b)
}

// GetBatchWithImages devuelve el batch y todas sus imágenes en un solo response.
func (h *BatchHandler) GetBatchWithImages(c *gin.Context) {
	batch, images, err := h.svc.GetBatchWithImages(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Batch not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"batch":  batch,
		"images": images,
		"total":  len(images),
	})
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

func (h *BatchHandler) SaveTransformations(c *gin.Context) {
	// Read raw body for debugging
	bodyBytes, _ := c.GetRawData()
	fmt.Printf("[DEBUG] Raw Transformations JSON: %s\n", string(bodyBytes))

	// Put it back for binding
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	batchUUID := c.Param("id")
	var transformations []struct {
		Name           string `json:"name"`
		Params         string `json:"params"`
		ExecutionOrder int    `json:"execution_order"`
	}

	if err := c.ShouldBindJSON(&transformations); err != nil {
		fmt.Printf("[DEBUG] Transformations Bind Error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domainTransformations := make([]image.BatchTransformation, 0, len(transformations))
	for _, t := range transformations {
		typeName := t.Name
		typeParams := t.Params

		// Handle Java's weird double-JSON encoding: {"name": "{\"name\": \"blur\", \"params\": \"...\"}"}
		if strings.HasPrefix(typeName, "{") {
			var nested struct {
				Name   string `json:"name"`
				Params string `json:"params"`
			}
			if err := json.Unmarshal([]byte(typeName), &nested); err == nil {
				if nested.Name != "" {
					typeName = nested.Name
				}
				// If the inner JSON has params and the outer one is empty or just "{}", use the inner ones
				if nested.Params != "" && (typeParams == "" || typeParams == "{}") {
					typeParams = nested.Params
				}
			}
		}

		typeID := utils.GetIDFromStatus(utils.TransformationTypes, typeName)

		domainTransformations = append(domainTransformations, image.BatchTransformation{
			BatchUUID:      batchUUID,
			TypeID:         typeID,
			Params:         typeParams,
			ExecutionOrder: t.ExecutionOrder,
		})
	}

	if err := h.svc.SaveTransformations(c.Request.Context(), batchUUID, domainTransformations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Transformations registered successfully"})
}
