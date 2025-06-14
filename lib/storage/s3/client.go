package s3

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/snowmerak/DraftStore/lib/storage"
)

var _ storage.Storage = (*Client)(nil)

type Client struct {
	client    *s3.Client
	presigner *s3.PresignClient
}

type ClientOptions struct {
	Region string
	Config *aws.Config
}

func NewClient(opts ClientOptions) (*Client, error) {
	var cfg aws.Config
	var err error

	if opts.Config != nil {
		cfg = *opts.Config
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return nil, err
		}
	}

	if opts.Region != "" {
		cfg.Region = opts.Region
	}

	client := s3.NewFromConfig(cfg)
	presigner := s3.NewPresignClient(client)

	return &Client{
		client:    client,
		presigner: presigner,
	}, nil
}

// CreateBucket implements storage.Storage.
func (c *Client) CreateBucket(ctx context.Context, bucketName string) error {
	_, err := c.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}

// DeleteBucket implements storage.Storage.
func (c *Client) DeleteBucket(ctx context.Context, bucketName string) error {
	_, err := c.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}

// ExistsBucket implements storage.Storage.
func (c *Client) ExistsBucket(ctx context.Context, bucketName string) (bool, error) {
	_, err := c.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "NotFound":
				return false, nil
			default:
				return false, err
			}
		}
		return false, err
	}
	return true, nil
}

// MakeGetPresignedURL implements storage.Storage.
func (c *Client) MakeGetPresignedURL(ctx context.Context, bucketName string, objectName string, ttl time.Duration) (string, error) {
	request, err := c.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = ttl
	})
	if err != nil {
		return "", err
	}
	return request.URL, nil
}

// MakeUploadPresignedURL implements storage.Storage.
func (c *Client) MakeUploadPresignedURL(ctx context.Context, bucketName string, objectName string, ttl time.Duration) (string, error) {
	request, err := c.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = ttl
	})
	if err != nil {
		return "", err
	}
	return request.URL, nil
}

// CleanupBucket implements storage.Storage.
func (c *Client) CleanupBucket(ctx context.Context, bucketName string, criteria time.Time, duration time.Duration) error {
	// List objects in the bucket
	paginator := s3.NewListObjectsV2Paginator(c.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	var objectsToDelete []string

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		for _, obj := range page.Contents {
			if obj.LastModified != nil {
				// Check if object meets criteria for deletion
				objTime := *obj.LastModified
				if objTime.Before(criteria) && time.Since(objTime) > duration {
					objectsToDelete = append(objectsToDelete, *obj.Key)
				}
			}
		}
	}

	// Delete objects in batches (S3 allows up to 1000 objects per batch)
	batchSize := 1000
	for i := 0; i < len(objectsToDelete); i += batchSize {
		end := i + batchSize
		if end > len(objectsToDelete) {
			end = len(objectsToDelete)
		}

		batch := objectsToDelete[i:end]
		var deleteObjects []types.ObjectIdentifier

		for _, key := range batch {
			deleteObjects = append(deleteObjects, types.ObjectIdentifier{
				Key: aws.String(key),
			})
		}

		if len(deleteObjects) > 0 {
			_, err := c.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(bucketName),
				Delete: &types.Delete{
					Objects: deleteObjects,
				},
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyObject implements storage.Storage.
func (c *Client) CopyObject(ctx context.Context, srcBucket string, srcObject string, dstBucket string, dstObject string) error {
	copySource := srcBucket + "/" + srcObject

	_, err := c.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(dstBucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(dstObject),
	})

	return err
}

// DeleteObject implements storage.Storage.
func (c *Client) DeleteObject(ctx context.Context, bucketName string, objectName string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})

	return err
}
