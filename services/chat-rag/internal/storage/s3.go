package storage

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// S3Config configures log archival without changing object keys.
type S3Config struct {
	Endpoint        string // host[:port]; protocol is selected by UseSSL
	Bucket          string
	AccessKey       string
	SecretKey       string
	UseSSL          bool
	SkipSSLVerify   bool
	Region          string // Empty uses us-east-1 for S3-compatible services.
	SkipBucketCheck bool
	// Nil means true; pointers distinguish omitted values from explicit false.
	ForcePathStyle         *bool
	DisableRequestChecksum *bool
}

type S3Storage struct {
	client *s3.Client
	bucket string
}

func enabledByDefault(value *bool) bool { return value == nil || *value }

// NewS3Storage checks the bucket and creates it only when it is missing,
// unless SkipBucketCheck is enabled.
func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	scheme := "http"
	if cfg.UseSSL {
		scheme = "https"
	}
	endpoint, err := url.Parse(scheme + "://" + cfg.Endpoint)
	if err != nil || endpoint.Hostname() == "" || endpoint.User != nil || endpoint.Path != "" || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return nil, fmt.Errorf("s3 storage: endpoint must be host[:port]")
	}
	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}
	awsCfg := aws.Config{
		Region:      region,
		Credentials: credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
	}
	if cfg.UseSSL && cfg.SkipSSLVerify {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // Explicit configuration for private/self-signed services.
		awsCfg.HTTPClient = &http.Client{Transport: transport}
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint.String())
		o.UsePathStyle = enabledByDefault(cfg.ForcePathStyle)
		if enabledByDefault(cfg.DisableRequestChecksum) {
			o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		} else {
			o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenSupported
		}
	})
	if !cfg.SkipBucketCheck {
		ctx := context.Background()
		_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(cfg.Bucket)})
		if err != nil {
			var responseErr *smithyhttp.ResponseError
			if !errors.As(err, &responseErr) || responseErr.HTTPStatusCode() != http.StatusNotFound {
				return nil, fmt.Errorf("s3 storage: failed to check bucket existence: %w", err)
			}
			input := &s3.CreateBucketInput{Bucket: aws.String(cfg.Bucket)}
			if region != "us-east-1" {
				input.CreateBucketConfiguration = &types.CreateBucketConfiguration{LocationConstraint: types.BucketLocationConstraint(region)}
			}
			if _, err := client.CreateBucket(ctx, input); err != nil {
				return nil, fmt.Errorf("s3 storage: failed to create bucket %q: %w", cfg.Bucket, err)
			}
		}
	}
	return &S3Storage{client: client, bucket: cfg.Bucket}, nil
}

// Write preserves the caller's key and bytes, including trailing newlines.
func (s *S3Storage) Write(key string, data []byte) (*WriteInfo, error) {
	info, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket), Key: aws.String(key),
		Body: bytes.NewReader(data), ContentLength: aws.Int64(int64(len(data))),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 storage: failed to write object %q: %w", key, err)
	}
	// MinIO returned unquoted ETags; preserve that contract for metrics consumers.
	return &WriteInfo{FilePath: key, ETag: strings.Trim(aws.ToString(info.ETag), "\"")}, nil
}

func (s *S3Storage) Close() error { return nil }
