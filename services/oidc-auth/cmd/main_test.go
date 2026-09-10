package main

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/zgsm-ai/oidc-auth/internal/config"
)

func TestNewHTTPClientExplicitProxy(t *testing.T) {
	cfg := &config.HTTPClientConfig{Timeout: time.Minute}
	client, err := newHTTPClient(cfg, "http://127.0.0.1:7890")
	if err != nil {
		t.Fatalf("newHTTPClient() error = %v", err)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T, want *http.Transport", client.Transport)
	}
	requestURL, _ := url.Parse("https://api.github.com")
	request := &http.Request{URL: requestURL}
	proxy, err := transport.Proxy(request)
	if err != nil {
		t.Fatalf("proxy lookup error = %v", err)
	}
	if proxy.String() != "http://127.0.0.1:7890" {
		t.Fatalf("proxy = %q, want %q", proxy, "http://127.0.0.1:7890")
	}
}

func TestNewHTTPClientRejectsInvalidProxy(t *testing.T) {
	_, err := newHTTPClient(&config.HTTPClientConfig{}, "://bad-proxy")
	if err == nil {
		t.Fatal("newHTTPClient() error = nil, want invalid proxy error")
	}
}

func TestNewHTTPClientRequiresConfig(t *testing.T) {
	_, err := newHTTPClient(nil, "")
	if err == nil {
		t.Fatal("newHTTPClient() error = nil, want missing config error")
	}
}
