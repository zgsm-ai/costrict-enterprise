package storage

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/viper"
	"github.com/zgsm-ai/chat-rag/internal/config"
)

func TestS3CompatibilityUpload(t *testing.T) {
	for _, secure := range []bool{false, true} {
		name := "HTTP"
		if secure {
			name = "HTTPS"
		}
		t.Run(name, func(t *testing.T) {
			var attempts atomic.Int32
			key := "2026-11/03/测试 user/20261103-120000_request_123.json"
			payload := "{\n  \"message\": \"日志\"\n}\n"
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := attempts.Add(1)
				if r.Method != http.MethodPut || r.URL.Path != "/log-bucket/"+key {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				body, err := io.ReadAll(r.Body)
				if err != nil || string(body) != payload {
					t.Errorf("payload changed: %q, %v", body, err)
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Error("content type changed")
				}
				if r.Header.Get("X-Amz-Trailer") != "" || r.Header.Get("X-Amz-Checksum-Crc32") != "" || strings.Contains(r.Header.Get("Content-Encoding"), "aws-chunked") {
					t.Error("unexpected checksum or chunk encoding")
				}
				if r.Header.Get("Authorization") == "" {
					t.Error("request is not signed")
				}
				if n == 1 {
					w.WriteHeader(500)
					return
				}
				w.Header().Set("ETag", `"test-etag"`)
			})
			var server *httptest.Server
			if secure {
				server = httptest.NewTLSServer(handler)
			} else {
				server = httptest.NewServer(handler)
			}
			defer server.Close()
			store, err := NewS3Storage(S3Config{Endpoint: strings.SplitN(server.URL, "://", 2)[1], Bucket: "log-bucket", AccessKey: "test", SecretKey: "test", UseSSL: secure, SkipSSLVerify: secure, SkipBucketCheck: true})
			if err != nil {
				t.Fatal(err)
			}
			info, err := store.Write(key, []byte(payload))
			if err != nil {
				t.Fatal(err)
			}
			if info.FilePath != key || info.ETag != "test-etag" {
				t.Fatalf("metadata changed: %+v", info)
			}
			if attempts.Load() != 2 {
				t.Fatalf("attempts=%d", attempts.Load())
			}
		})
	}
}

func TestS3BucketInitialization(t *testing.T) {
	for _, tc := range []struct {
		name                        string
		status                      int
		skip, wantCreate, wantError bool
	}{
		{name: "exists", status: 200},
		{name: "missing", status: 404, wantCreate: true},
		{name: "forbidden", status: 403, wantError: true},
		{name: "skipped", status: 403, skip: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var heads, creates atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/log-bucket" {
					t.Errorf("unexpected bucket path %s", r.URL.Path)
				}
				switch r.Method {
				case http.MethodHead:
					heads.Add(1)
					w.WriteHeader(tc.status)
				case http.MethodPut:
					creates.Add(1)
					w.WriteHeader(200)
				default:
					t.Errorf("unexpected method %s", r.Method)
				}
			}))
			defer server.Close()
			_, err := NewS3Storage(S3Config{Endpoint: strings.TrimPrefix(server.URL, "http://"), Bucket: "log-bucket", AccessKey: "test", SecretKey: "test", SkipBucketCheck: tc.skip})
			if (err != nil) != tc.wantError {
				t.Fatalf("error=%v", err)
			}
			if (creates.Load() == 1) != tc.wantCreate {
				t.Errorf("create requests=%d", creates.Load())
			}
			if tc.skip && heads.Load() != 0 {
				t.Error("bucket check was not skipped")
			}
		})
	}
}

func TestS3ConfigCompatibilityDefaults(t *testing.T) {
	for _, tc := range []struct {
		name, yaml string
		want       bool
	}{
		{"omitted", "log:\n  s3:\n    bucket: test\n", true},
		{"enabled", "log:\n  s3:\n    forcePathStyle: true\n    disableRequestChecksum: true\n", true},
		{"disabled", "log:\n  s3:\n    forcePathStyle: false\n    disableRequestChecksum: false\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := viper.New()
			v.SetConfigType("yaml")
			if err := v.ReadConfig(strings.NewReader(tc.yaml)); err != nil {
				t.Fatal(err)
			}
			var cfg struct {
				Log struct{ S3 config.LogS3Config }
			}
			if err := v.Unmarshal(&cfg); err != nil {
				t.Fatal(err)
			}
			store, err := NewS3Storage(S3Config{Endpoint: "localhost:9000", Bucket: "test", SkipBucketCheck: true, ForcePathStyle: cfg.Log.S3.ForcePathStyle, DisableRequestChecksum: cfg.Log.S3.DisableRequestChecksum})
			if err != nil {
				t.Fatal(err)
			}
			opts := store.client.Options()
			if opts.UsePathStyle != tc.want {
				t.Errorf("path style=%v", opts.UsePathStyle)
			}
			required := opts.RequestChecksumCalculation == aws.RequestChecksumCalculationWhenRequired
			if required != tc.want {
				t.Errorf("checksum policy=%v", opts.RequestChecksumCalculation)
			}
		})
	}
}

func TestS3ChecksumCanBeEnabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		if r.Header.Get("X-Amz-Checksum-Crc32") == "" {
			t.Error("explicit false did not enable checksum")
		}
	}))
	defer server.Close()
	disabled := false
	store, err := NewS3Storage(S3Config{Endpoint: strings.TrimPrefix(server.URL, "http://"), Bucket: "test", AccessKey: "test", SecretKey: "test", SkipBucketCheck: true, DisableRequestChecksum: &disabled})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Write("log.json", []byte("{}\n")); err != nil {
		t.Fatal(err)
	}
}
