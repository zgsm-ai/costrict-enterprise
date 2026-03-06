package rpc

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"costrict-keeper/internal/logger"
)

// httpClient HTTP客户端实现
type httpClient struct {
	config *HTTPConfig
	client *http.Client
}

// NewHTTPClient 创建HTTP客户端实例
/**
 * Create new HTTP client for Unix socket communication
 * @param {HTTPConfig} config - HTTP client configuration
 * @returns {HTTPClient} HTTP client interface
 * @returns {error} Error if client creation fails
 * @description
 * - Creates HTTP client configured for Unix socket communication
 * - Initializes custom transport for Unix socket connection
 * - Sets default configuration if none provided
 * - Configures timeout and connection settings
 * @throws
 * - Configuration validation errors
 * - Transport initialization errors
 * @example
 * config := DefaultHTTPConfig()
 * client, err := NewHTTPClient(config)
 * if err != nil {
 *     log.Fatal(err)
 * }
 */
func NewHTTPClient(config *HTTPConfig) HTTPClient {
	if config == nil {
		config = DefaultHTTPConfig()
	}

	hc := &httpClient{
		config: config,
	}

	// 初始化transport，但不立即连接
	transport := &http.Transport{
		// 其他配置可以在这里设置
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial(config.Network, config.Address)
		},
	}

	hc.client = &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return hc
}

/**
 * Send GET request to server via Unix socket
 * @param {string} path - API endpoint path
 * @param {map[string]interface{}} params - Query parameters
 * @returns {interface{}} Response data
 * @returns {error} Error if request fails
 * @description
 * - Constructs URL with base URL and path
 * - Adds query parameters to request
 * - Establishes Unix socket connection if not connected
 * - Sends HTTP GET request and parses response
 * - Handles connection errors and timeouts
 * @throws
 * - URL construction errors
 * - Connection establishment errors
 * - HTTP request errors
 * - Response parsing errors
 * @example
 * result, err := client.Get("/api/components", map[string]interface{}{
 *     "status": "active",
 * })
 * if err != nil {
 *     log.Fatal(err)
 * }
 */
func (c *httpClient) Get(path string, params map[string]interface{}) (*HTTPResponse, error) {
	url, err := buildURL(c.config.BaseURL, path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	logger.Debugf("Sending GET request to %s", url)

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	httpResp, err := deserializeResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response: %w", err)
	}

	return httpResp, nil
}

/**
 * Send POST request to server via Unix socket
 * @param {string} path - API endpoint path
 * @param {interface{}} data - Request body data
 * @returns {interface{}} Response data
 * @returns {error} Error if request fails
 * @description
 * - Constructs URL with base URL and path
 * - Serializes request body to JSON
 * - Establishes Unix socket connection if not connected
 * - Sends HTTP POST request and parses response
 * - Handles connection errors and timeouts
 * @throws
 * - URL construction errors
 * - Data serialization errors
 * - Connection establishment errors
 * - HTTP request errors
 * - Response parsing errors
 * @example
 * data := map[string]interface{}{
 *     "name": "test",
 *     "value": 123,
 * }
 * result, err := client.Post("/api/components", data)
 * if err != nil {
 *     log.Fatal(err)
 * }
 */
func (c *httpClient) Post(path string, data interface{}) (*HTTPResponse, error) {
	url, err := buildURL(c.config.BaseURL, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	body, err := serializeData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}

	logger.Debugf("Sending POST request to %s", url)

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	httpResp, err := deserializeResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response: %w", err)
	}

	return httpResp, nil
}

/**
 * Send PUT request to server via Unix socket
 * @param {string} path - API endpoint path
 * @param {interface{}} data - Request body data
 * @returns {interface{}} Response data
 * @returns {error} Error if request fails
 * @description
 * - Constructs URL with base URL and path
 * - Serializes request body to JSON
 * - Establishes Unix socket connection if not connected
 * - Sends HTTP PUT request and parses response
 * - Handles connection errors and timeouts
 * @throws
 * - URL construction errors
 * - Data serialization errors
 * - Connection establishment errors
 * - HTTP request errors
 * - Response parsing errors
 * @example
 * data := map[string]interface{}{
 *     "name": "updated",
 *     "value": 456,
 * }
 * result, err := client.Put("/api/components/1", data)
 * if err != nil {
 *     log.Fatal(err)
 * }
 */
func (c *httpClient) Put(path string, data interface{}) (*HTTPResponse, error) {
	url, err := buildURL(c.config.BaseURL, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	body, err := serializeData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}

	logger.Debugf("Sending PUT request to %s", url)

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "PUT", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	httpResp, err := deserializeResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response: %w", err)
	}

	return httpResp, nil
}

/**
 * Send PATCH request to server via Unix socket
 * @param {string} path - API endpoint path
 * @param {interface{}} data - Request body data
 * @returns {interface{}} Response data
 * @returns {error} Error if request fails
 * @description
 * - Constructs URL with base URL and path
 * - Serializes request body to JSON
 * - Establishes Unix socket connection if not connected
 * - Sends HTTP PATCH request and parses response
 * - Handles connection errors and timeouts
 * @throws
 * - URL construction errors
 * - Data serialization errors
 * - Connection establishment errors
 * - HTTP request errors
 * - Response parsing errors
 * @example
 * data := map[string]interface{}{
 *     "value": 789,
 * }
 * result, err := client.Patch("/api/components/1", data)
 * if err != nil {
 *     log.Fatal(err)
 * }
 */
func (c *httpClient) Patch(path string, data interface{}) (*HTTPResponse, error) {
	url, err := buildURL(c.config.BaseURL, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	body, err := serializeData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}

	logger.Debugf("Sending PATCH request to %s", url)

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	httpResp, err := deserializeResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response: %w", err)
	}

	return httpResp, nil
}

/**
 * Send DELETE request to server via Unix socket
 * @param {string} path - API endpoint path
 * @param {map[string]interface{}} params - Query parameters
 * @returns {interface{}} Response data
 * @returns {error} Error if request fails
 * @description
 * - Constructs URL with base URL and path
 * - Adds query parameters to request
 * - Establishes Unix socket connection if not connected
 * - Sends HTTP DELETE request and parses response
 * - Handles connection errors and timeouts
 * @throws
 * - URL construction errors
 * - Connection establishment errors
 * - HTTP request errors
 * - Response parsing errors
 * @example
 * result, err := client.Delete("/api/components/1", nil)
 * if err != nil {
 *     log.Fatal(err)
 * }
 */
func (c *httpClient) Delete(path string, params map[string]interface{}) (*HTTPResponse, error) {
	url, err := buildURL(c.config.BaseURL, path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	logger.Debugf("Sending DELETE request to %s", url)

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	httpResp, err := deserializeResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response: %w", err)
	}

	return httpResp, nil
}

/**
 * Close HTTP client connection
 * @returns {error} Error if closing fails
 * @description
 * - Closes HTTP client and transport
 * - Resets connection state
 * - Cleans up resources
 * @throws
 * - Resource cleanup errors
 * @example
 * defer client.Close()
 */
func (c *httpClient) Close() error {
	if c.client != nil {
		c.client.CloseIdleConnections()
	}

	// if c.transport != nil {
	// 	c.transport.CloseIdleConnections()
	// }

	logger.Debugf("HTTP client connection closed")
	return nil
}
