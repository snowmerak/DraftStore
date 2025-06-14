package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/snowmerak/DraftStore/lib/service/cleaner"
	"github.com/snowmerak/DraftStore/lib/storage/s3"
)

type Config struct {
	BucketName     string
	AWSRegion      string
	ObjectLifetime time.Duration
}

func loadConfig() *Config {
	cfg := &Config{
		BucketName:     getEnv("BUCKET_NAME", "main"),
		AWSRegion:      getEnv("AWS_REGION", "us-east-1"),
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

func getDurationEnv(key string, defaultValue int64) time.Duration {
	if value := os.Getenv(key); value != "" {
		if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
			return time.Duration(seconds)
		}
	}
	return time.Duration(defaultValue)
}

func main() {
	// Load configuration
	cfg := loadConfig()

	log.Printf("Starting DraftStore cleanup job...")
	log.Printf("Configuration: Bucket=%s, Region=%s, ObjectLifetime=%v",
		cfg.BucketName, cfg.AWSRegion, cfg.ObjectLifetime)

	// Initialize S3 storage client
	storageClient, err := s3.NewClient(s3.ClientOptions{
		Region: cfg.AWSRegion,
	})
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
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
