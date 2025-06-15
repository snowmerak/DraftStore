package draft

import (
	"context"
	"fmt"
	"time"

	"github.com/snowmerak/DraftStore/lib/storage"
	"github.com/snowmerak/DraftStore/lib/util/logger"
)

const (
	DefaultDraftBucketSuffix = "-draft"
)

type Service struct {
	bucketName  string
	draftBucket string
	storage     storage.Storage
	uploadTTL   time.Duration
	downloadTTL time.Duration
}

type ServiceOptions struct {
	BucketName  string
	Storage     storage.Storage
	UploadTTL   time.Duration
	DownloadTTL time.Duration
}

func NewService(opts ServiceOptions) (*Service, error) {
	log := logger.GetServiceLogger("draft-service")

	service := &Service{
		storage:     opts.Storage,
		bucketName:  opts.BucketName,
		draftBucket: opts.BucketName + DefaultDraftBucketSuffix,
		uploadTTL:   opts.UploadTTL,
		downloadTTL: opts.DownloadTTL,
	}

	log.Info().
		Str("bucket_name", service.bucketName).
		Str("draft_bucket", service.draftBucket).
		Dur("upload_ttl", service.uploadTTL).
		Dur("download_ttl", service.downloadTTL).
		Msg("Draft service initialized")

	return service, nil
}

func (s *Service) CreateDraftBucket(ctx context.Context) error {
	log := logger.GetServiceLogger("draft-service").With().
		Str("operation", "create_draft_bucket").
		Str("draft_bucket", s.draftBucket).
		Str("main_bucket", s.bucketName).
		Logger()

	log.Info().Msg("Starting bucket creation operation")

	// Check if draft bucket exists
	exists, err := s.storage.ExistsBucket(ctx, s.draftBucket)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to check if draft bucket exists")
		return fmt.Errorf("failed to check if draft bucket exists: %w", err)
	}

	if exists {
		log.Info().Msg("Draft bucket already exists")
	} else {
		log.Info().Msg("Creating draft bucket")
		if err := s.storage.CreateBucket(ctx, s.draftBucket); err != nil {
			log.Error().
				Err(err).
				Msg("Failed to create draft bucket")
			return fmt.Errorf("failed to create draft bucket %s: %w", s.draftBucket, err)
		}

		logger.LogStateChange("create", "bucket", s.draftBucket, nil, map[string]interface{}{
			"bucket_name": s.draftBucket,
			"type":        "draft",
		})
	}

	// Check if main bucket exists
	exists, err = s.storage.ExistsBucket(ctx, s.bucketName)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to check if main bucket exists")
		return fmt.Errorf("failed to check if main bucket exists: %w", err)
	}

	if exists {
		log.Info().Msg("Main bucket already exists")
	} else {
		log.Info().Msg("Creating main bucket")
		if err := s.storage.CreateBucket(ctx, s.bucketName); err != nil {
			log.Error().
				Err(err).
				Msg("Failed to create main bucket")
			return fmt.Errorf("failed to create main bucket %s: %w", s.bucketName, err)
		}

		logger.LogStateChange("create", "bucket", s.bucketName, nil, map[string]interface{}{
			"bucket_name": s.bucketName,
			"type":        "main",
		})
	}

	log.Info().Msg("Bucket creation operation completed successfully")
	return nil
}

func (s *Service) GetUploadURL(ctx context.Context, objectName string) (string, error) {
	log := logger.GetServiceLogger("draft-service").With().
		Str("operation", "get_upload_url").
		Str("object_name", objectName).
		Str("bucket", s.bucketName).
		Dur("ttl", s.uploadTTL).
		Logger()

	log.Info().Msg("Generating upload URL")

	url, err := s.storage.MakeUploadPresignedURL(ctx, s.bucketName, objectName, s.uploadTTL)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to generate upload URL")
		return "", fmt.Errorf("failed to get upload URL: %w", err)
	}

	log.Info().
		Str("url_length", fmt.Sprintf("%d", len(url))).
		Msg("Upload URL generated successfully")
	return url, nil
}

func (s *Service) GetDownloadURL(ctx context.Context, objectName string) (string, error) {
	log := logger.GetServiceLogger("draft-service").With().
		Str("operation", "get_download_url").
		Str("object_name", objectName).
		Str("bucket", s.bucketName).
		Dur("ttl", s.downloadTTL).
		Logger()

	log.Info().Msg("Generating download URL")

	url, err := s.storage.MakeGetPresignedURL(ctx, s.bucketName, objectName, s.downloadTTL)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to generate download URL")
		return "", fmt.Errorf("failed to get download URL: %w", err)
	}

	log.Info().
		Str("url_length", fmt.Sprintf("%d", len(url))).
		Msg("Download URL generated successfully")
	return url, nil
}

func (s *Service) ConfirmUpload(ctx context.Context, objectName string) error {
	log := logger.GetServiceLogger("draft-service").With().
		Str("operation", "confirm_upload").
		Str("object_name", objectName).
		Str("source_bucket", s.draftBucket).
		Str("dest_bucket", s.bucketName).
		Logger()

	log.Info().Msg("Starting upload confirmation process")

	// Copy object from draft bucket to main bucket
	if err := s.storage.CopyObject(ctx, s.draftBucket, objectName, s.bucketName, objectName); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to copy object from draft to main bucket")
		return fmt.Errorf("failed to confirm upload: %w", err)
	}

	log.Info().Msg("Object copied successfully, now deleting from draft bucket")

	// Delete object from draft bucket
	if err := s.storage.DeleteObject(ctx, s.draftBucket, objectName); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to delete object from draft bucket after confirmation")
		return fmt.Errorf("failed to delete draft object after confirmation: %w", err)
	}

	// Log the state change
	logger.LogStateChange("confirm_upload", "object", objectName,
		map[string]interface{}{
			"location": "draft_bucket",
			"bucket":   s.draftBucket,
		},
		map[string]interface{}{
			"location": "main_bucket",
			"bucket":   s.bucketName,
		})

	log.Info().Msg("Upload confirmation completed successfully")
	return nil
}
