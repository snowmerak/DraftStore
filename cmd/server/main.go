package main

import (
	"context"
	"fmt"
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
	log := logger.GetServiceLogger("server")

	// Load configuration
	cfg := loadConfig()

	// Log startup configuration
	logger.LogStartup("server", map[string]interface{}{
		"storage_type": cfg.StorageType,
		"bucket_name":  cfg.BucketName,
		"grpc_port":    cfg.GRPCPort,
		"http_port":    cfg.HTTPPort,
		"upload_ttl":   cfg.UploadTTL.String(),
		"download_ttl": cfg.DownloadTTL.String(),
	})

	switch cfg.StorageType {
	case "s3":
		log.Info().
			Str("region", cfg.AWSRegion).
			Msg("Using AWS S3 storage backend")
	case "minio":
		log.Info().
			Str("endpoint", cfg.MinIOEndpoint).
			Str("access_key", cfg.MinIOAccessKey).
			Bool("use_ssl", cfg.MinIOUseSSL).
			Str("region", cfg.MinIORegion).
			Msg("Using MinIO storage backend")
	default:
		log.Fatal().
			Str("storage_type", cfg.StorageType).
			Msg("Unsupported storage type")

		// Fallback to default configuration
		log.Warn().Msg("Falling back to default configuration")
		cfg.StorageType = "s3"
		cfg.AWSRegion = "us-east-1"
		cfg.BucketName = "main"
		cfg.GRPCPort = "50051"
		cfg.HTTPPort = "8080"
		cfg.UploadTTL = 3600 * time.Second
		cfg.DownloadTTL = 3600 * time.Second

		log.Info().
			Str("storage_type", cfg.StorageType).
			Str("bucket", cfg.BucketName).
			Str("grpc_port", cfg.GRPCPort).
			Str("http_port", cfg.HTTPPort).
			Dur("upload_ttl", cfg.UploadTTL).
			Dur("download_ttl", cfg.DownloadTTL).
			Msg("Applied default configuration")
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

	// Initialize draft service
	log.Info().Msg("Initializing draft service")
	draftService, err := draft.NewService(draft.ServiceOptions{
		BucketName:  cfg.BucketName,
		Storage:     storageClient,
		UploadTTL:   cfg.UploadTTL,
		DownloadTTL: cfg.DownloadTTL,
	})
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to create draft service")
	}
	log.Info().Msg("Draft service initialized successfully")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start gRPC server
	log.Info().
		Str("port", cfg.GRPCPort).
		Msg("Starting gRPC server")
	grpcServer := startGRPCServer(cfg.GRPCPort, draftService)
	defer grpcServer.GracefulStop()

	// Start HTTP server
	log.Info().
		Str("port", cfg.HTTPPort).
		Msg("Starting HTTP server")
	httpServer := startHTTPServer(cfg.HTTPPort, draftService)
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error().
				Err(err).
				Msg("HTTP server shutdown error")
		}
	}()

	log.Info().
		Str("grpc_port", cfg.GRPCPort).
		Str("http_port", cfg.HTTPPort).
		Msg("All servers started successfully")

	// Wait for interrupt signal
	waitForShutdown(ctx)

	// Log shutdown information
	logger.LogShutdown("server", time.Since(startTime))
}

func startGRPCServer(port string, draftService *draft.Service) *grpc.Server {
	log := logger.GetServiceLogger("grpc-server")

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Create and register draft service
	draftGRPCServer := grpcController.NewServer(grpcController.ServerOptions{
		DraftService: draftService,
		Address:      ":" + port,
	})

	draftv1.RegisterDraftServiceServer(grpcServer, draftGRPCServer)
	log.Info().
		Str("port", port).
		Msg("gRPC service registered")

	// Start listening
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("port", port).
			Msg("Failed to listen on gRPC port")
	}

	// Start server in goroutine
	go func() {
		log.Info().
			Str("address", ":"+port).
			Msg("gRPC server starting to serve")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to serve gRPC")
		}
	}()

	log.Info().
		Str("port", port).
		Msg("gRPC server started successfully")
	return grpcServer
}

func startHTTPServer(port string, draftService *draft.Service) *http.Server {
	log := logger.GetServiceLogger("http-server")

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

	log.Info().Msg("HTTP middleware configured")

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

	log.Info().
		Str("address", ":"+port).
		Dur("read_timeout", 30*time.Second).
		Dur("write_timeout", 30*time.Second).
		Dur("idle_timeout", 120*time.Second).
		Msg("HTTP server configured")

	// Start server in goroutine
	go func() {
		log.Info().
			Str("address", ":"+port).
			Msg("HTTP server starting to serve")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().
				Err(err).
				Msg("Failed to serve HTTP")
		}
	}()

	log.Info().
		Str("port", port).
		Msg("HTTP server started successfully")
	return httpServer
}

func waitForShutdown(ctx context.Context) {
	log := logger.GetServiceLogger("shutdown")

	// Create channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Info().
			Str("signal", sig.String()).
			Msg("Received shutdown signal")
	case <-ctx.Done():
		log.Info().Msg("Context cancelled - initiating shutdown")
	}
}
