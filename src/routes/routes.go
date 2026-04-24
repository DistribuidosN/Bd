package routes

import (
	"net/http"

	app_handlers "enfok_bd/src/handlers/http"

	"github.com/gin-gonic/gin"
)

// Handlers agrupa todos los handlers de entrada del API.
type Handlers struct {
	Image   *app_handlers.ImageHandler
	Batch   *app_handlers.BatchHandler
	Node    *app_handlers.NodeHandler
	Log     *app_handlers.LogHandler
	Metrics *app_handlers.MetricsHandler
}

// Setup registra todas las rutas de la API v1.
func Setup(router *gin.Engine, h *Handlers) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "enfok_bd"})
	})

	api := router.Group("/api/v1")
	{
		// Imágenes
		api.POST("/images", h.Image.UploadImage)
		api.GET("/images/:id", h.Image.GetImage)
		api.GET("/images/:id/download", h.Image.DownloadImage)
		api.PATCH("/images/:id/status", h.Image.UpdateStatus)
		api.PATCH("/images/:id/result", h.Image.UpdateResultPath)

		// Batches
		api.GET("/batches", h.Batch.ListBatches)
		api.POST("/batches/upload", h.Image.UploadBatch)
		api.GET("/batches/:id", h.Batch.GetBatch)
		api.GET("/batches/:id/images", h.Image.GetBatchImages)
		api.PATCH("/batches/:id/status", h.Batch.UpdateStatus)
		api.POST("/batches/:id/transformations", h.Batch.SaveTransformations)
		api.GET("/batches/:id/progress", h.Batch.GetProgress)

		// Nodos
		api.POST("/nodes", h.Node.Register)
		api.GET("/nodes", h.Node.List)
		api.GET("/nodes/:node_id", h.Node.GetNode)
		api.POST("/nodes/:node_id/heartbeat", h.Node.Heartbeat)
		api.PATCH("/nodes/:node_id/status", h.Node.UpdateStatus)

		// Logs
		api.POST("/logs", h.Log.CreateLog)
		api.GET("/logs/:image_uuid", h.Log.GetLogsByImage)

		// Métricas
		api.POST("/metrics", h.Metrics.CreateMetrics)
		api.GET("/metrics/:node_id", h.Metrics.GetMetricsByNode)
		api.GET("/metrics/image/:image_uuid", h.Metrics.GetMetricsByImage)

		// Usuarios (en ImageHandler por simplicidad de dependencias)
		api.GET("/users/:user_uuid/statistics", h.Image.GetUserStatistics)
		api.GET("/users/:user_uuid/activity", h.Image.GetUserActivity)
	}
}

