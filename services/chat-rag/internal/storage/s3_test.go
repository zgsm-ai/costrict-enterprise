package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Compile-time check: S3Storage must satisfy StorageBackend.
var _ StorageBackend = (*S3Storage)(nil)

// ---------------------------------------------------------------------------
// Unit tests — run without a live MinIO instance.
// ---------------------------------------------------------------------------

func TestS3Storage_CloseReturnsNil(t *testing.T) {
	// Close is a documented no-op; it must return nil even when the underlying
	// S3 client is absent (struct constructed directly, not via NewS3Storage).
	s := &S3Storage{bucket: "test-bucket"}
	if err := s.Close(); err != nil {
		t.Errorf("Close() returned unexpected error: %v", err)
	}
}

func TestS3Config_Fields(t *testing.T) {
	cfg := S3Config{
		Endpoint:  "minio.example.com:9000",
		Bucket:    "my-bucket",
		AccessKey: "AKID",
		SecretKey: "SECRET",
		UseSSL:    true,
		Region:    "us-east-1",
	}

	// Verify all fields round-trip correctly.
	checks := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Endpoint", cfg.Endpoint, "minio.example.com:9000"},
		{"Bucket", cfg.Bucket, "my-bucket"},
		{"AccessKey", cfg.AccessKey, "AKID"},
		{"SecretKey", cfg.SecretKey, "SECRET"},
		{"UseSSL", cfg.UseSSL, true},
		{"Region", cfg.Region, "us-east-1"},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("S3Config.%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Integration tests — require a running MinIO instance.
//
// These tests are skipped automatically unless MINIO_ENDPOINT is set and
// -short is not passed. Run them with:
//
//   MINIO_ENDPOINT=localhost:9000 \
//   MINIO_ACCESS_KEY=minioadmin \
//   MINIO_SECRET_KEY=minioadmin \
//   go test ./internal/storage/... -v -run Integration
//
// ---------------------------------------------------------------------------

func skipIfNoMinIO(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	if os.Getenv("MINIO_ENDPOINT") == "" {
		t.Skip("skipping integration test: MINIO_ENDPOINT not set")
	}
}

func TestS3Storage_Integration_WriteAndVerify(t *testing.T) {
	skipIfNoMinIO(t)

	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")

	bucket := "s3storage-integration-test"

	cfg := S3Config{
		Endpoint:  endpoint,
		Bucket:    bucket,
		AccessKey: accessKey,
		SecretKey: secretKey,
		UseSSL:    false,
		Region:    "",
	}

	s, err := NewS3Storage(cfg)
	if err != nil {
		t.Fatalf("NewS3Storage() returned error: %v", err)
	}
	defer s.Close()

	// Write an object.
	key := "integration-test/hello.json"
	data := []byte(`{"msg":"hello from integration test"}`)

	if _, err := s.Write(key, data); err != nil {
		t.Fatalf("Write(%q) returned error: %v", key, err)
	}

	obj, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	if err != nil {
		t.Fatalf("GetObject(%q) returned error: %v", key, err)
	}
	defer obj.Body.Close()

	got, err := io.ReadAll(obj.Body)
	if err != nil {
		t.Fatalf("failed to read object body: %v", err)
	}

	if !bytes.Equal(got, data) {
		t.Errorf("object content mismatch:\n  got:  %q\n  want: %q", got, data)
	}

	// Cleanup: remove the test object (best-effort).
	_, _ = s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
}
