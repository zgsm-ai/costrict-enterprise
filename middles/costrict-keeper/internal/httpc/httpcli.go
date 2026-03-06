package httpc

import (
	"crypto/tls"
	"net/http"
	"time"
)

// Using shared HTTP client instances with connection pooling
var httpClient *http.Client

/**
 * Get shared HTTP client for server communications
 * @returns {*http.Client} Returns the shared HTTP client instance
 * @description
 * - Provides access to the shared HTTP client for all modules
 * - Ensures consistent HTTP configuration across the application
 * - Used by tunnel_manager, upgrade utilities, and log service
 * - Enables connection pooling and resource optimization
 * @example
 * server := NewServer(cfg)
 * client := server.GetClient()
 * resp, err := client.Get("https://example.com")
 */
func GetClient() *http.Client {
	if httpClient == nil {
		InitClient(5, 30, true)
	}
	return httpClient
}

func InitClient(idleConns, idleConnTimeout int, insecureSkipVerify bool) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
		MaxIdleConns:    idleConns,
		IdleConnTimeout: time.Duration(idleConnTimeout) * time.Second,
	}
	httpClient = &http.Client{Transport: transport}
}

/**
 * Close idle HTTP connections to cleanup resources
 * @description
 * - Closes all idle HTTP connections in the shared transport
 * - Should be called during server shutdown to release resources
 * - Prevents connection leaks and ensures clean shutdown
 * - Safe to call multiple times
 */
func CloseConnections() {
	if httpClient != nil && httpClient.Transport != nil {
		httpClient.CloseIdleConnections()
	}
}
