package draft

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
	return &Service{
		storage:     opts.Storage,
		bucketName:  opts.BucketName,
		draftBucket: opts.BucketName + DefaultDraftBucketSuffix,
		uploadTTL:   opts.UploadTTL,
		downloadTTL: opts.DownloadTTL,
	}, nil
}

func (s *Service) CreateDraftBucket(ctx context.Context) error {
	exists, err := s.storage.ExistsBucket(ctx, s.draftBucket)
	if err != nil {
		return fmt.Errorf("failed to check if draft bucket exists: %w", err)
	}
	if exists {
		return nil // Bucket already exists, no need to create it
	}

	if err := s.storage.CreateBucket(ctx, s.draftBucket); err != nil {
		return fmt.Errorf("failed to create draft bucket %s: %w", s.draftBucket, err)
	}

	exists, err = s.storage.ExistsBucket(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check if main bucket exists: %w", err)
	}
	if exists {
		return nil // Main bucket already exists, no need to create it
	}

	if err := s.storage.CreateBucket(ctx, s.bucketName); err != nil {
		return fmt.Errorf("failed to create main bucket %s: %w", s.bucketName, err)
	}

	return nil
}

func (s *Service) GetUploadURL(ctx context.Context, objectName string) (string, error) {
	url, err := s.storage.MakeUploadPresignedURL(ctx, s.bucketName, objectName, s.uploadTTL)
	if err != nil {
		return "", fmt.Errorf("failed to get upload URL: %w", err)
	}

	return url, nil
}

func (s *Service) GetDownloadURL(ctx context.Context, objectName string) (string, error) {
	url, err := s.storage.MakeGetPresignedURL(ctx, s.bucketName, objectName, s.downloadTTL)
	if err != nil {
		return "", fmt.Errorf("failed to get download URL: %w", err)
	}

	return url, nil
}

func (s *Service) ConfirmUpload(ctx context.Context, objectName string) error {
	if err := s.storage.CopyObject(ctx, s.draftBucket, objectName, s.bucketName, objectName); err != nil {
		return fmt.Errorf("failed to confirm upload: %w", err)
	}

	if err := s.storage.DeleteObject(ctx, s.draftBucket, objectName); err != nil {
		return fmt.Errorf("failed to delete draft object after confirmation: %w", err)
	}

	return nil
}
