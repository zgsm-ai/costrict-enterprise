package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// HTTPClientConfig defines the configuration for HTTP client
type HTTPClientConfig struct {
	Timeout time.Duration
}

// HTTPClient represents a generic HTTP client
type HTTPClient struct {
	endpoint   string
	httpClient *http.Client
}

// NewHTTPClient creates a new HTTP client instance
func NewHTTPClient(endpoint string, config HTTPClientConfig) *HTTPClient {
	if config.Timeout == 0 {
		config.Timeout = 3 * time.Second
	}

	return &HTTPClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Request represents a generic HTTP request
type Request struct {
	Method        string            `json:"-"`
	Path          string            `json:"-"`
	QueryParams   map[string]string `json:"-"`
	Headers       map[string]string `json:"-"`
	Body          interface{}       `json:"-"`
	Authorization string            `json:"-"`
}

// DoRequest executes an HTTP request and returns the response
func (c *HTTPClient) DoRequest(ctx context.Context, req Request) (*http.Response, error) {
	// Create URL
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint URL: %w", err)
	}

	// Add path if provided
	if req.Path != "" {
		u.Path = req.Path
	}

	// Add query parameters if provided
	if len(req.QueryParams) > 0 {
		params := url.Values{}
		for key, value := range req.QueryParams {
			params.Add(key, value)
		}
		u.RawQuery = params.Encode()
	}

	// Prepare request body
	var body io.Reader
	if req.Body != nil {
		if req.Method == http.MethodGet || req.Method == http.MethodHead {
			// For GET/HEAD requests, marshal body to query parameters if it's a struct
			if bodyMap, err := structToMap(req.Body); err == nil {
				params := u.Query()
				for key, value := range bodyMap {
					params.Add(key, fmt.Sprintf("%v", value))
				}
				u.RawQuery = params.Encode()
			}
		} else {
			// For other methods, marshal body to JSON
			reqBody, err := json.Marshal(req.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			body = bytes.NewReader(reqBody)
		}
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}

	// Set default content type for POST/PUT/PATCH requests
	if (req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch) &&
		req.Headers["Content-Type"] == "" && req.Body != nil {
		req.Headers["Content-Type"] = "application/json"
	}

	// Set authorization header if provided
	if req.Authorization != "" {
		req.Headers["Authorization"] = req.Authorization
	}

	// Apply headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

// structToMap converts a struct to a map[string]interface{}
func structToMap(obj interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}
