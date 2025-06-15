package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/snowmerak/DraftStore/lib/service/cleaner"
	"github.com/snowmerak/DraftStore/lib/storage"
	"github.com/snowmerak/DraftStore/lib/storage/minio"
	"github.com/snowmerak/DraftStore/lib/storage/s3"
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
	switch cfg.StorageType {
	case "s3":
		return s3.NewClient(s3.ClientOptions{
			Region: cfg.AWSRegion,
		})
	case "minio":
		return minio.NewClient(minio.ClientOptions{
			Endpoint:        cfg.MinIOEndpoint,
			AccessKeyID:     cfg.MinIOAccessKey,
			SecretAccessKey: cfg.MinIOSecretKey,
			UseSSL:          cfg.MinIOUseSSL,
			Region:          cfg.MinIORegion,
		})
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.StorageType)
	}
}

func main() {
	// Load configuration
	cfg := loadConfig()

	log.Printf("Starting DraftStore cleanup job...")
	log.Printf("Configuration: StorageType=%s, Bucket=%s, ObjectLifetime=%v",
		cfg.StorageType, cfg.BucketName, cfg.ObjectLifetime)

	if cfg.StorageType == "s3" {
		log.Printf("Using AWS S3 with region: %s", cfg.AWSRegion)
	} else if cfg.StorageType == "minio" {
		log.Printf("Using MinIO with endpoint: %s", cfg.MinIOEndpoint)
	}

	// Initialize storage client
	storageClient, err := createStorageClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create storage client: %v", err)
	}

	// Initialize cleaner service
	cleanerService, err := cleaner.NewService(cleaner.ServiceOptions{
		BucketName:     cfg.BucketName,
		ObjectLifetime: cfg.ObjectLifetime,
		Storage:        storageClient,
	})
	if err != nil {
		log.Fatalf("Failed to create cleaner service: %v", err)
	}

	// Run cleanup once and exit (designed for Kubernetes Job)
	log.Printf("Running cleanup job...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := cleanerService.CleanupDrafts(ctx); err != nil {
		log.Fatalf("Cleanup failed: %v", err)
	}

	log.Printf("Cleanup completed successfully")
	log.Printf("Job finished, exiting...")
}
