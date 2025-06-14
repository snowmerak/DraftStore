package storage

import (
	"context"
	"time"
)

type Storage interface {
	CreateBucket(ctx context.Context, bucketName string) error
	DeleteBucket(ctx context.Context, bucketName string) error
	ExistsBucket(ctx context.Context, bucketName string) (bool, error)
	MakeUploadPresignedURL(ctx context.Context, bucketName, objectName string, ttl time.Duration) (string, error)
	MakeGetPresignedURL(ctx context.Context, bucketName, objectName string, ttl time.Duration) (string, error)
	CopyObject(ctx context.Context, srcBucket, srcObject, dstBucket, dstObject string) error
	DeleteObject(ctx context.Context, bucketName, objectName string) error
	CleanupBucket(ctx context.Context, bucketName string, criteria time.Time, duration time.Duration) error
}
