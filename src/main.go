package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	http_handlers "enfok_bd/src/handlers/http"
	"enfok_bd/src/infrastructure/config"
	"enfok_bd/src/infrastructure/db"
	"enfok_bd/src/infrastructure/repository"
	"enfok_bd/src/routes"
	"enfok_bd/src/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Initialize Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. Setup Context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 3. Load App Configuration
	cfg := config.LoadConfig()

	// 4. Connect to Infrastructure
	clients, err := db.Connect(ctx, cfg, logger)
	if err != nil {
		logger.Error("Critical error initializing infrastructure", "error", err)
		os.Exit(1)
	}
	defer clients.DB.Close()

	// 5. Initialize Components (DI)
	imgRepo := repository.NewPostgresImageRepository(clients.DB)
	batchRepo := repository.NewPostgresBatchRepository(clients.DB)
	nodeRepo := repository.NewPostgresNodeRepository(clients.DB)
	logRepo := repository.NewPostgresLogRepository(clients.DB)
	metricRepo := repository.NewPostgresMetricRepository(clients.DB)
	storageRepo := repository.NewMinioStorageRepository(clients.MinioInternal, clients.MinioExternal, cfg.Minio.Bucket)

	imgSvc := service.NewImageService(imgRepo, storageRepo, batchRepo)
	batchSvc := service.NewBatchService(batchRepo, imgRepo, storageRepo)
	logSvc := service.NewLogService(logRepo)
	metricSvc := service.NewMetricsService(metricRepo)
	nodeSvc := service.NewNodeService(nodeRepo)

	h := &routes.Handlers{
		Image:   http_handlers.NewImageHandler(imgSvc),
		Batch:   http_handlers.NewBatchHandler(batchSvc),
		Node:    http_handlers.NewNodeHandler(nodeSvc),
		Log:     http_handlers.NewLogHandler(logSvc),
		Metrics: http_handlers.NewMetricsHandler(metricSvc),
	}

	// 5.1 Start background tasks
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logger.Info("Running cleanup for old exports...")
				if err := storageRepo.DeleteOldExports(context.Background(), 24*time.Hour); err != nil {
					logger.Error("Error cleaning up old exports", "error", err)
				}
			}
		}
	}()

	// 6. Setup Router & Routes
	router := gin.Default()
	routes.Setup(router, h)

	// 7. Start HTTP Server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		logger.Info("Starting server", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server gracefully...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Error("Graceful shutdown failed", "error", err)
	}

	logger.Info("Enfok API stopped")
}
