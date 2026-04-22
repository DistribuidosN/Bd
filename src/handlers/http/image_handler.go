package http

import (
	"io"
	"net/http"

	"enfok_bd/src/domain/ports/driving"

	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	svc driving.ImageServicePort
}

func NewImageHandler(s driving.ImageServicePort) *ImageHandler {
	return &ImageHandler{svc: s}
}

func (h *ImageHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image provided"})
		return
	}
	defer file.Close()

	userUUID := c.PostForm("user_uuid")
	if userUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_uuid is required"})
		return
	}

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
		return
	}

	imageUUID, err := h.svc.ProcessUpload(c.Request.Context(), userUUID, header.Filename, content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Image uploaded successfully", "image_uuid": imageUUID})
}

func (h *ImageHandler) GetImage(c *gin.Context) {
	img, err := h.svc.GetImage(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}
	c.JSON(http.StatusOK, img)
}

func (h *ImageHandler) UpdateStatus(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
}
