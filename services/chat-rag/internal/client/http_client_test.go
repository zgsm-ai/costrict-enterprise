package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	endpoint := "http://localhost:8002/api"
	config := HTTPClientConfig{
		Timeout: 5 * time.Second,
	}

	client := NewHTTPClient(endpoint, config)

	if client == nil {
		t.Fatal("NewHTTPClient returned nil")
	}

	if client.httpClient == nil {
		t.Fatal("HTTP client is nil")
	}

	if client.httpClient.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", client.httpClient.Timeout)
	}
}

func TestNewHTTPClient_DefaultTimeout(t *testing.T) {
	endpoint := "http://localhost:8002/api"
	config := HTTPClientConfig{} // Zero timeout

	client := NewHTTPClient(endpoint, config)

	if client.httpClient.Timeout != 3*time.Second {
		t.Errorf("Expected default timeout 3s, got %v", client.httpClient.Timeout)
	}
}

func TestHTTPClient_DoRequest_POST(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Verify authorization header
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got '%s'", r.Header.Get("Authorization"))
		}

		// Verify request body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
			return
		}

		if reqBody["test"] != "value" {
			t.Errorf("Expected request body to contain test=value, got %+v", reqBody)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "success",
			"data":    "test response",
		})
	}))
	defer server.Close()

	config := HTTPClientConfig{
		Timeout: 3 * time.Second,
	}
	client := NewHTTPClient(server.URL, config)

	req := Request{
		Method:        http.MethodPost,
		Authorization: "Bearer test-token",
		Body:          map[string]interface{}{"test": "value"},
	}

	ctx := context.Background()
	resp, err := client.DoRequest(ctx, req)

	if err != nil {
		t.Fatalf("DoRequest failed: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTPClient_DoRequest_GET(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		// Verify query parameters
		if r.URL.Query().Get("param1") != "value1" {
			t.Errorf("Expected param1=value1, got '%s'", r.URL.Query().Get("param1"))
		}

		if r.URL.Query().Get("param2") != "value2" {
			t.Errorf("Expected param2=value2, got '%s'", r.URL.Query().Get("param2"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "success",
			"data":    "test response",
		})
	}))
	defer server.Close()

	config := HTTPClientConfig{
		Timeout: 3 * time.Second,
	}
	client := NewHTTPClient(server.URL, config)

	req := Request{
		Method: http.MethodGet,
		QueryParams: map[string]string{
			"param1": "value1",
			"param2": "value2",
		},
	}

	ctx := context.Background()
	resp, err := client.DoRequest(ctx, req)

	if err != nil {
		t.Fatalf("DoRequest failed: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTPClient_DoRequest_GET_WithBody(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		// Verify query parameters (body should be converted to query params for GET)
		if r.URL.Query().Get("test") != "value" {
			t.Errorf("Expected test=value in query params, got '%s'", r.URL.Query().Get("test"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "success",
			"data":    "test response",
		})
	}))
	defer server.Close()

	config := HTTPClientConfig{
		Timeout: 3 * time.Second,
	}
	client := NewHTTPClient(server.URL, config)

	req := Request{
		Method: http.MethodGet,
		Body:   map[string]interface{}{"test": "value"},
	}

	ctx := context.Background()
	resp, err := client.DoRequest(ctx, req)

	if err != nil {
		t.Fatalf("DoRequest failed: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTPClient_DoRequest_WithPath(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify path
		if r.URL.Path != "/api/test" {
			t.Errorf("Expected path /api/test, got '%s'", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "success",
			"data":    "test response",
		})
	}))
	defer server.Close()

	config := HTTPClientConfig{
		Timeout: 3 * time.Second,
	}
	client := NewHTTPClient(server.URL, config)

	req := Request{
		Method: http.MethodGet,
		Path:   "/api/test",
	}

	ctx := context.Background()
	resp, err := client.DoRequest(ctx, req)

	if err != nil {
		t.Fatalf("DoRequest failed: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTPClient_DoRequest_HTTPError(t *testing.T) {
	// Mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	config := HTTPClientConfig{
		Timeout: 3 * time.Second,
	}
	client := NewHTTPClient(server.URL, config)

	req := Request{
		Method: http.MethodGet,
	}

	ctx := context.Background()
	resp, err := client.DoRequest(ctx, req)

	// DoRequest should not return error for HTTP status errors
	// It should return the response and let the caller handle the status
	if err != nil {
		t.Fatalf("DoRequest failed: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
}

func TestHTTPClient_DoRequest_ContextCancellation(t *testing.T) {
	// Mock server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "success",
			"data":    "test response",
		})
	}))
	defer server.Close()

	config := HTTPClientConfig{
		Timeout: 3 * time.Second,
	}
	client := NewHTTPClient(server.URL, config)

	req := Request{
		Method: http.MethodGet,
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.DoRequest(ctx, req)

	if err == nil {
		t.Fatal("Expected context cancellation error, got nil")
	}

	expectedError := "failed to execute request"
	if !strings.HasPrefix(err.Error(), expectedError) {
		t.Errorf("Expected error to start with '%s', got '%s'", expectedError, err.Error())
	}
}

func TestStructToMap(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	testStruct := TestStruct{
		Name:  "test",
		Value: 42,
	}

	result, err := structToMap(testStruct)
	if err != nil {
		t.Fatalf("structToMap failed: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("Expected name 'test', got '%v'", result["name"])
	}

	if result["value"] != 42.0 { // JSON numbers are unmarshaled as float64
		t.Errorf("Expected value 42, got %v", result["value"])
	}
}

func TestHTTPClient_DoRequest_InvalidURL(t *testing.T) {
	config := HTTPClientConfig{
		Timeout: 3 * time.Second,
	}
	client := NewHTTPClient("invalid-url", config)

	req := Request{
		Method: http.MethodGet,
	}

	ctx := context.Background()
	_, err := client.DoRequest(ctx, req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedError := "failed to execute request"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestHTTPClient_DoRequest_CustomHeaders(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify custom headers
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			t.Errorf("Expected X-Custom-Header 'custom-value', got '%s'", r.Header.Get("X-Custom-Header"))
		}

		if r.Header.Get("X-Another-Header") != "another-value" {
			t.Errorf("Expected X-Another-Header 'another-value', got '%s'", r.Header.Get("X-Another-Header"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "success",
			"data":    "test response",
		})
	}))
	defer server.Close()

	config := HTTPClientConfig{
		Timeout: 3 * time.Second,
	}
	client := NewHTTPClient(server.URL, config)

	req := Request{
		Method: http.MethodGet,
		Headers: map[string]string{
			"X-Custom-Header":  "custom-value",
			"X-Another-Header": "another-value",
		},
	}

	ctx := context.Background()
	resp, err := client.DoRequest(ctx, req)

	if err != nil {
		t.Fatalf("DoRequest failed: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
