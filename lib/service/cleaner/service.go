package cleaner

import (
	"context"
	"fmt"
	"time"

	"github.com/snowmerak/DraftStore/lib/storage"
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
	return &Service{
		bucketName:     opts.BucketName,
		draftBucket:    opts.BucketName + DefaultDraftBucketSuffix,
		objectLifetime: opts.ObjectLifetime,
		storage:        opts.Storage,
	}, nil
}

func (s *Service) CleanupDrafts(ctx context.Context) error {
	// Get the current time
	now := time.Now()

	// Perform the cleanup operation
	if err := s.storage.CleanupBucket(ctx, s.draftBucket, now, s.objectLifetime); err != nil {
		return fmt.Errorf("failed to cleanup drafts in bucket %s: %w", s.draftBucket, err)
	}

	return nil
}
