package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DBConfig
	Minio    MinioConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type DBConfig struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
}

type MinioConfig struct {
	URL      string
	User     string
	Password string
	Bucket   string
	SSL      bool
}

// LoadConfig carga configuración desde .env o variables del sistema.
// Hace panic si faltan variables críticas.
func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Host: getEnvOrDefault("SERVER_HOST", "localhost"),
			Port: mustEnv("SERVER_PORT"),
		},
		Database: DBConfig{
			DBUser:     mustEnv("DB_USER"),
			DBPassword: mustEnv("DB_PASSWORD"),
			DBHost:     mustEnv("DB_HOST"),
			DBPort:     mustEnv("DB_PORT"),
			DBName:     mustEnv("DB_NAME"),
		},
		Minio: MinioConfig{
			URL:      mustEnv("MINIO_URL"),
			User:     mustEnv("MINIO_USER"),
			Password: mustEnv("MINIO_PWD"),
			Bucket:   mustEnv("MINIO_BUCKET"),
			SSL:      getEnvOrDefault("MINIO_SSL", "false") == "true",
		},
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("variable de entorno requerida no definida: %s", key))
	}
	return v
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
