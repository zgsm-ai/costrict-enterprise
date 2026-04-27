package storage

import (
	"bytes"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Config holds the configuration required to connect to an S3-compatible
// object storage service (e.g., AWS S3, MinIO, Ceph RGW).
type S3Config struct {
	Endpoint  string // Host and optional port, e.g. "s3.amazonaws.com" or "minio:9000"
	Bucket    string // Target bucket name; created automatically if it does not exist
	AccessKey string
	SecretKey string
	UseSSL    bool   // Whether to use HTTPS for the connection
	Region    string // Bucket region, e.g. "us-east-1"; may be empty for MinIO
	// SkipBucketCheck disables the BucketExists / MakeBucket calls during init.
	// Useful when the credentials lack s3:ListBucket / s3:CreateBucket but the
	// bucket is already provisioned out-of-band. Defaults to false (check enabled).
	SkipBucketCheck bool
}

// S3Storage implements StorageBackend by writing objects to an S3-compatible
// object storage service via the minio-go v7 SDK.
// A single minio.Client is created at construction time and reused for all
// writes. The target bucket is auto-created during initialization if it does
// not already exist.
type S3Storage struct {
	client *minio.Client
	bucket string
}

// NewS3Storage creates an S3Storage backed by the given S3Config.
// It initialises the minio client, checks whether the target bucket exists,
// and creates it when absent. Any error during client creation or bucket
// setup is returned immediately.
func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("s3 storage: failed to create client: %w", err)
	}

	if !cfg.SkipBucketCheck {
		ctx := context.Background()

		exists, err := client.BucketExists(ctx, cfg.Bucket)
		if err != nil {
			return nil, fmt.Errorf("s3 storage: failed to check bucket existence: %w", err)
		}
		if !exists {
			err = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{
				Region: cfg.Region,
			})
			if err != nil {
				return nil, fmt.Errorf("s3 storage: failed to create bucket %q: %w", cfg.Bucket, err)
			}
		}
	}

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Write persists data as an object with the given key in the configured bucket.
// The object is stored with content type "application/json".
// ChatLog JSON payloads are typically 1-50 KB so single-part upload is sufficient.
func (s *S3Storage) Write(key string, data []byte) error {
	reader := bytes.NewReader(data)
	_, err := s.client.PutObject(context.Background(), s.bucket, key, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		return fmt.Errorf("s3 storage: failed to write object %q: %w", key, err)
	}
	return nil
}

// Close is a no-op for S3Storage because the minio-go client does not hold
// persistent connections or resources that require explicit teardown.
func (s *S3Storage) Close() error {
	return nil
}
