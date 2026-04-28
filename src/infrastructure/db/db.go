package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"enfok_bd/src/infrastructure/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Clients struct {
	DB            *sqlx.DB
	MinioInternal *minio.Client
	MinioExternal *minio.Client
}

// Connect abre la conexión a Postgres y MinIO, configurando los pools.
func Connect(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Clients, error) {
	// 1. Initialize Postgres with sqlx
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.DBUser,
		cfg.Database.DBPassword,
		cfg.Database.DBHost,
		cfg.Database.DBPort,
		cfg.Database.DBName,
	)

	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("error abriendo conexión a postgres: %w", err)
	}

	// CONFIGURACIÓN DEL POOL
	db.SetMaxOpenConns(25)                 // máximo de conexiones abiertas
	db.SetMaxIdleConns(10)                 // conexiones en espera (reutilizables)
	db.SetConnMaxIdleTime(5 * time.Minute) // tiempo máximo idle
	db.SetConnMaxLifetime(1 * time.Hour)   // vida máxima de una conexión

	// Verificación con contexto
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("no se pudo conectar a postgres en %s:%s: %w",
			cfg.Database.DBHost, cfg.Database.DBPort, err)
	}

	logger.Info("conexión a postgres establecida con pool",
		"host", cfg.Database.DBHost,
		"port", cfg.Database.DBPort,
		"database", cfg.Database.DBName,
	)

	// 2. Initialize MinIO Internal (localhost:9999 para operaciones locales)
	internalClient, err := minio.New(cfg.Minio.URL, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.User, cfg.Minio.Password, ""),
		Secure: cfg.Minio.SSL,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create internal minio client: %w", err)
	}

	// Ensure bucket exists
	exists, err := internalClient.BucketExists(ctx, cfg.Minio.Bucket)
	if err != nil {
		return nil, fmt.Errorf("unable to check bucket: %w", err)
	}
	if !exists {
		if err := internalClient.MakeBucket(ctx, cfg.Minio.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("unable to create bucket: %w", err)
		}
		logger.Info("Bucket creado", "bucket", cfg.Minio.Bucket)
	}

	// Ensure exports bucket exists
	exportsBucket := "exports"
	exportsExists, err := internalClient.BucketExists(ctx, exportsBucket)
	if err != nil {
		return nil, fmt.Errorf("unable to check exports bucket: %w", err)
	}
	if !exportsExists {
		if err := internalClient.MakeBucket(ctx, exportsBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("unable to create exports bucket: %w", err)
		}
		logger.Info("Bucket de exportaciones creado", "bucket", exportsBucket)
	}

	logger.Info("conexión a MinIO Interna establecida", "url", cfg.Minio.URL)

	// 3. Initialize MinIO External (ngrok para firmas públicas)
	ngrokURL := "913d-181-55-22-220.ngrok-free.app"
	externalClient, err := minio.New(ngrokURL, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.User, cfg.Minio.Password, ""),
		Secure: true, // ngrok usa https
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create external minio client: %w", err)
	}

	return &Clients{
		DB:            db,
		MinioInternal: internalClient,
		MinioExternal: externalClient,
	}, nil
}
