package minio

import (
	"context"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/snowmerak/DraftStore/lib/storage"
	"github.com/snowmerak/DraftStore/lib/util/logger"
)

var _ storage.Storage = (*Client)(nil)

type Client struct {
	client *minio.Client
}

type ClientOptions struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Region          string
	Config          *minio.Options
}

func NewClient(opts ClientOptions) (*Client, error) {
	log := logger.GetServiceLogger("minio-storage")

	log.Info().
		Str("endpoint", opts.Endpoint).
		Str("region", opts.Region).
		Bool("use_ssl", opts.UseSSL).
		Msg("Initializing MinIO storage client")

	var client *minio.Client
	var err error

	if opts.Config != nil {
		client, err = minio.New(opts.Endpoint, opts.Config)
	} else {
		client, err = minio.New(opts.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(opts.AccessKeyID, opts.SecretAccessKey, ""),
			Secure: opts.UseSSL,
			Region: opts.Region,
		})
	}

	if err != nil {
		log.Error().
			Err(err).
			Str("endpoint", opts.Endpoint).
			Msg("Failed to create MinIO client")
		return nil, err
	}

	log.Info().
		Str("endpoint", opts.Endpoint).
		Str("region", opts.Region).
		Msg("MinIO storage client initialized successfully")

	return &Client{
		client: client,
	}, nil
}

// CreateBucket implements storage.Storage.
func (c *Client) CreateBucket(ctx context.Context, bucketName string) error {
	return c.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
}

// DeleteBucket implements storage.Storage.
func (c *Client) DeleteBucket(ctx context.Context, bucketName string) error {
	return c.client.RemoveBucket(ctx, bucketName)
}

// ExistsBucket implements storage.Storage.
func (c *Client) ExistsBucket(ctx context.Context, bucketName string) (bool, error) {
	return c.client.BucketExists(ctx, bucketName)
}

// MakeGetPresignedURL implements storage.Storage.
func (c *Client) MakeGetPresignedURL(ctx context.Context, bucketName string, objectName string, ttl time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedGetObject(ctx, bucketName, objectName, ttl, nil)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

// MakeUploadPresignedURL implements storage.Storage.
func (c *Client) MakeUploadPresignedURL(ctx context.Context, bucketName string, objectName string, ttl time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedPutObject(ctx, bucketName, objectName, ttl)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

// CopyObject implements storage.Storage.
func (c *Client) CopyObject(ctx context.Context, srcBucket string, srcObject string, dstBucket string, dstObject string) error {
	srcOpts := minio.CopySrcOptions{
		Bucket: srcBucket,
		Object: srcObject,
	}
	dstOpts := minio.CopyDestOptions{
		Bucket: dstBucket,
		Object: dstObject,
	}
	_, err := c.client.CopyObject(ctx, dstOpts, srcOpts)
	return err
}

// DeleteObject implements storage.Storage.
func (c *Client) DeleteObject(ctx context.Context, bucketName string, objectName string) error {
	return c.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}

// CleanupBucket implements storage.Storage.
func (c *Client) CleanupBucket(ctx context.Context, bucketName string, criteria time.Time, duration time.Duration) error {
	objectCh := c.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Recursive:    true,
		WithMetadata: true,
	})

	var objectsToDelete []minio.ObjectInfo

	for object := range objectCh {
		if object.Err != nil {
			return object.Err
		}

		// Check if object meets criteria for deletion
		if object.LastModified.Before(criteria) && time.Since(object.LastModified) > duration {
			objectsToDelete = append(objectsToDelete, object)
		}
	}

	// Delete objects
	for _, obj := range objectsToDelete {
		err := c.client.RemoveObject(ctx, bucketName, obj.Key, minio.RemoveObjectOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
