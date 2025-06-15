package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/snowmerak/DraftStore/lib/service/cleaner"
	"github.com/snowmerak/DraftStore/lib/storage"
	"github.com/snowmerak/DraftStore/lib/storage/minio"
	"github.com/snowmerak/DraftStore/lib/storage/s3"
	"github.com/snowmerak/DraftStore/lib/util/logger"
)

type Config struct {
	StorageType string
	BucketName  string
	// AWS S3 Configuration
	AWSRegion string
	// MinIO Configuration
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOUseSSL    bool
	MinIORegion    string
	// Cleanup Configuration
	ObjectLifetime time.Duration
}

func loadConfig() *Config {
	cfg := &Config{
		StorageType: getEnv("STORAGE_TYPE", "s3"),
		BucketName:  getEnv("BUCKET_NAME", "main"),
		// AWS S3 Configuration
		AWSRegion: getEnv("AWS_REGION", "us-east-1"),
		// MinIO Configuration
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOUseSSL:    getBoolEnv("MINIO_USE_SSL", false),
		MinIORegion:    getEnv("MINIO_REGION", "us-east-1"),
		// Cleanup Configuration
		ObjectLifetime: getDurationEnv("OBJECT_LIFETIME", 86400) * time.Second,
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if value == "true" || value == "1" || value == "yes" {
			return true
		}
		if value == "false" || value == "0" || value == "no" {
			return false
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue int64) time.Duration {
	if value := os.Getenv(key); value != "" {
		if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
			return time.Duration(seconds)
		}
	}
	return time.Duration(defaultValue)
}

func createStorageClient(cfg *Config) (storage.Storage, error) {
	log := logger.GetServiceLogger("storage")

	switch cfg.StorageType {
	case "s3":
		log.Info().
			Str("region", cfg.AWSRegion).
			Msg("Creating S3 storage client")
		return s3.NewClient(s3.ClientOptions{
			Region: cfg.AWSRegion,
		})
	case "minio":
		log.Info().
			Str("endpoint", cfg.MinIOEndpoint).
			Str("region", cfg.MinIORegion).
			Bool("use_ssl", cfg.MinIOUseSSL).
			Msg("Creating MinIO storage client")
		return minio.NewClient(minio.ClientOptions{
			Endpoint:        cfg.MinIOEndpoint,
			AccessKeyID:     cfg.MinIOAccessKey,
			SecretAccessKey: cfg.MinIOSecretKey,
			UseSSL:          cfg.MinIOUseSSL,
			Region:          cfg.MinIORegion,
		})
	default:
		log.Error().
			Str("storage_type", cfg.StorageType).
			Msg("Unsupported storage type")
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.StorageType)
	}
}

func main() {
	startTime := time.Now()
	log := logger.GetServiceLogger("cleanup-job")

	// Load configuration
	cfg := loadConfig()

	// Log startup information
	logger.LogStartup("cleanup-job", map[string]interface{}{
		"storage_type":    cfg.StorageType,
		"bucket_name":     cfg.BucketName,
		"object_lifetime": cfg.ObjectLifetime.String(),
	})

	if cfg.StorageType == "s3" {
		log.Info().
			Str("region", cfg.AWSRegion).
			Msg("Using AWS S3 storage backend")
	} else if cfg.StorageType == "minio" {
		log.Info().
			Str("endpoint", cfg.MinIOEndpoint).
			Msg("Using MinIO storage backend")
	}

	// Initialize storage client
	log.Info().Msg("Initializing storage client")
	storageClient, err := createStorageClient(cfg)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("storage_type", cfg.StorageType).
			Msg("Failed to create storage client")
	}
	log.Info().
		Str("storage_type", cfg.StorageType).
		Msg("Storage client initialized successfully")

	// Initialize cleaner service
	log.Info().Msg("Initializing cleaner service")
	cleanerService, err := cleaner.NewService(cleaner.ServiceOptions{
		BucketName:     cfg.BucketName,
		ObjectLifetime: cfg.ObjectLifetime,
		Storage:        storageClient,
	})
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to create cleaner service")
	}
	log.Info().Msg("Cleaner service initialized successfully")

	// Run cleanup once and exit (designed for Kubernetes Job)
	log.Info().Msg("Starting cleanup operation")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := cleanerService.CleanupDrafts(ctx); err != nil {
		log.Fatal().
			Err(err).
			Msg("Cleanup operation failed")
	}

	log.Info().
		Dur("duration", time.Since(startTime)).
		Msg("Cleanup operation completed successfully")

	// Log shutdown information
	logger.LogShutdown("cleanup-job", time.Since(startTime))
}
