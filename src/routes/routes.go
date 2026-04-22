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
		api.PATCH("/images/:id/status", h.Image.UpdateStatus)

		// Batches
		api.GET("/batches/:id", h.Batch.GetBatch)
		api.PATCH("/batches/:id/status", h.Batch.UpdateStatus)

		// Nodos
		api.POST("/nodes", h.Node.Register)
		api.POST("/nodes/:node_id/heartbeat", h.Node.Heartbeat)
		api.GET("/nodes", h.Node.List)

		// Logs
		api.POST("/logs", h.Log.CreateLog)
		api.GET("/logs/:image_uuid", h.Log.GetLogsByImage)

		// Métricas
		api.POST("/metrics", h.Metrics.CreateMetrics)
		api.GET("/metrics/:node_id", h.Metrics.GetMetricsByNode)
	}
}
