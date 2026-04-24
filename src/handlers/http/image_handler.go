package http

import (
	"io"
	"net/http"

	"strconv"
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
	// Soporte para registro por JSON (sin bytes)
	if c.ContentType() == "application/json" {
		var body struct {
			UserUUID  string `json:"user_uuid" binding:"required"`
			BatchUUID string `json:"batch_uuid" binding:"required"`
			ImageUUID string `json:"image_uuid" binding:"required"`
			FileName  string `json:"file_name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := h.svc.RegisterImage(c.Request.Context(), body.UserUUID, body.BatchUUID, body.ImageUUID, body.FileName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Image registered successfully"})
		return
	}

	// Lógica original: multipart
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

func (h *ImageHandler) UpdateResultPath(c *gin.Context) {
	// 1. Leer el NodeID del Header
	nodeID := c.GetHeader("X-Node-Id")
	if nodeID == "" {
		nodeID = "unknown"
	}

	// 2. Leer bytes crudos directamente del body
	content, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	defer c.Request.Body.Close()

	if len(content) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty body"})
		return
	}

	// 3. Llamar al servicio con el nodeId
	resultPath, err := h.svc.UpdateResult(c.Request.Context(), c.Param("id"), content, nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Result path and NodeID updated successfully",
		"result_path": resultPath,
	})
}

// UploadBatch acepta múltiples archivos bajo el key "images" en multipart/form-data
// O acepta un JSON para registrar el batch record.
func (h *ImageHandler) UploadBatch(c *gin.Context) {
	// Soporte para registro de Batch por JSON
	if c.ContentType() == "application/json" {
		var body struct {
			UserUUID  string `json:"user_uuid" binding:"required"`
			BatchUUID string `json:"batch_uuid" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := h.svc.RegisterBatch(c.Request.Context(), body.UserUUID, body.BatchUUID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Batch registered successfully"})
		return
	}

	// Lógica original: multipart
	userUUID := c.PostForm("user_uuid")
	if userUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_uuid is required"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
		return
	}

	fileHeaders := form.File["images"]
	if len(fileHeaders) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one image is required under key 'images'"})
		return
	}

	var batchFiles []driving.BatchFile
	for _, fh := range fileHeaders {
		f, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file: " + fh.Filename})
			return
		}
		content, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file: " + fh.Filename})
			return
		}
		batchFiles = append(batchFiles, driving.BatchFile{
			FileName: fh.Filename,
			Content:  content,
		})
	}

	batchUUID, imageUUIDs, err := h.svc.UploadBatch(c.Request.Context(), userUUID, batchFiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Batch uploaded successfully",
		"batch_uuid":  batchUUID,
		"image_uuids": imageUUIDs,
		"total":       len(imageUUIDs),
	})
}

func (h *ImageHandler) GetUserStatistics(c *gin.Context) {
	stats, err := h.svc.GetUserStatistics(c.Request.Context(), c.Param("user_uuid"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *ImageHandler) GetUserActivity(c *gin.Context) {
	activity, err := h.svc.GetUserActivity(c.Request.Context(), c.Param("user_uuid"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, activity)
}

func (h *ImageHandler) GetBatchImages(c *gin.Context) {
	batchUUID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	result, err := h.svc.GetBatchImagesPaginated(c.Request.Context(), batchUUID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ImageHandler) DownloadImage(c *gin.Context) {
	uuid := c.Param("id")
	content, fileName, err := h.svc.DownloadImage(c.Request.Context(), uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Detect content type
	contentType := http.DetectContentType(content)

	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Data(http.StatusOK, contentType, content)
}

