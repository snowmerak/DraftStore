package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"

	draftv1 "github.com/snowmerak/DraftStore/gen/draft/v1"
	grpcController "github.com/snowmerak/DraftStore/lib/controller/grpc"
	webapiController "github.com/snowmerak/DraftStore/lib/controller/webapi"
	"github.com/snowmerak/DraftStore/lib/service/draft"
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
	// Server Configuration
	GRPCPort    string
	HTTPPort    string
	UploadTTL   time.Duration
	DownloadTTL time.Duration
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
		// Server Configuration
		GRPCPort:    getEnv("GRPC_PORT", "50051"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		UploadTTL:   getDurationEnv("UPLOAD_TTL", 3600) * time.Second,
		DownloadTTL: getDurationEnv("DOWNLOAD_TTL", 3600) * time.Second,
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

	log.Printf("Starting DraftStore server...")
	log.Printf("Configuration: StorageType=%s, Bucket=%s, gRPC=%s, HTTP=%s",
		cfg.StorageType, cfg.BucketName, cfg.GRPCPort, cfg.HTTPPort)

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

	// Initialize draft service
	draftService, err := draft.NewService(draft.ServiceOptions{
		BucketName:  cfg.BucketName,
		Storage:     storageClient,
		UploadTTL:   cfg.UploadTTL,
		DownloadTTL: cfg.DownloadTTL,
	})
	if err != nil {
		log.Fatalf("Failed to create draft service: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start gRPC server
	grpcServer := startGRPCServer(cfg.GRPCPort, draftService)
	defer grpcServer.GracefulStop()

	// Start HTTP server
	httpServer := startHTTPServer(cfg.HTTPPort, draftService)
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	log.Printf("Servers started successfully")
	log.Printf("gRPC server listening on :%s", cfg.GRPCPort)
	log.Printf("HTTP server listening on :%s", cfg.HTTPPort)

	// Wait for interrupt signal
	waitForShutdown(ctx)

	log.Printf("Shutting down servers...")
}

func startGRPCServer(port string, draftService *draft.Service) *grpc.Server {
	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Create and register draft service
	draftGRPCServer := grpcController.NewServer(grpcController.ServerOptions{
		DraftService: draftService,
		Address:      ":" + port,
	})

	draftv1.RegisterDraftServiceServer(grpcServer, draftGRPCServer)

	// Start listening
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	// Start server in goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	return grpcServer
}

func startHTTPServer(port string, draftService *draft.Service) *http.Server {
	// Create router with middleware
	router := chi.NewRouter()

	// Add middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(middleware.Heartbeat("/health"))

	// Add CORS headers
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Create web API server
	webAPIServer := webapiController.NewServer(webapiController.ServerOptions{
		Router:       router,
		Address:      ":" + port,
		DraftService: draftService,
	})

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      webAPIServer.GetRouter(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()

	return httpServer
}

func waitForShutdown(ctx context.Context) {
	// Create channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	case <-ctx.Done():
		log.Printf("Context cancelled")
	}
}
