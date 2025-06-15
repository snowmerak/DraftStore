package cleaner

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
	bucketName     string
	draftBucket    string
	objectLifetime time.Duration
	storage        storage.Storage
}

type ServiceOptions struct {
	BucketName     string
	ObjectLifetime time.Duration
	Storage        storage.Storage
}

func NewService(opts ServiceOptions) (*Service, error) {
	log := logger.GetServiceLogger("cleaner-service")

	service := &Service{
		bucketName:     opts.BucketName,
		draftBucket:    opts.BucketName + DefaultDraftBucketSuffix,
		objectLifetime: opts.ObjectLifetime,
		storage:        opts.Storage,
	}

	log.Info().
		Str("bucket_name", service.bucketName).
		Str("draft_bucket", service.draftBucket).
		Dur("object_lifetime", service.objectLifetime).
		Msg("Cleaner service initialized")

	return service, nil
}

func (s *Service) CleanupDrafts(ctx context.Context) error {
	log := logger.GetServiceLogger("cleaner-service").With().
		Str("operation", "cleanup_drafts").
		Str("bucket", s.draftBucket).
		Dur("object_lifetime", s.objectLifetime).
		Logger()

	// Get the current time
	now := time.Now()
	cutoffTime := now.Add(-s.objectLifetime)

	log.Info().
		Time("current_time", now).
		Time("cutoff_time", cutoffTime).
		Msg("Starting cleanup operation")

	// Perform the cleanup operation
	if err := s.storage.CleanupBucket(ctx, s.draftBucket, cutoffTime, s.objectLifetime); err != nil {
		log.Error().
			Err(err).
			Msg("Cleanup operation failed")
		return fmt.Errorf("failed to cleanup drafts in bucket %s: %w", s.draftBucket, err)
	}

	// Log the state change
	logger.LogStateChange("cleanup", "bucket", s.draftBucket,
		map[string]interface{}{
			"status": "before_cleanup",
		},
		map[string]interface{}{
			"status":      "after_cleanup",
			"cutoff_time": cutoffTime,
		})

	log.Info().Msg("Cleanup operation completed successfully")
	return nil
}
