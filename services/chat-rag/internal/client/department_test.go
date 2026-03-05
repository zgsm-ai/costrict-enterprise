package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zgsm-ai/chat-rag/internal/model"
)

func TestGetDepartment(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate request path
		if r.URL.Path != "/123456" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		response := `{
			"code": 200,
			"data": {
				"dept_1": "a-dept",
				"dept_1_id": "1",
				"dept_2": "b-dept",
				"dept_2_id": "2",
				"dept_3": "c-dept",
				"dept_3_id": "3",
				"dept_4": "d-dept",
				"dept_4_id": "4"
			},
			"message": null,
			"success": true
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer ts.Close()

	// Create client
	client := NewDepartmentClient(ts.URL + "/")
	client.SetTimeout(time.Second * 2)

	t.Run("successful request", func(t *testing.T) {
		deptInfo, err := client.GetDepartment("123456")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expected := model.DepartmentInfo{
			Level1Dept: "a-dept",
			Level2Dept: "b-dept",
			Level3Dept: "c-dept",
			Level4Dept: "d-dept",
		}

		fmt.Printf("deptInfo: %+v\n", deptInfo)

		if *deptInfo != expected {
			t.Errorf("Expected %+v, got %+v", expected, *deptInfo)
		}
	})

	t.Run("cache hit", func(t *testing.T) {
		// First call, should get from API
		_, err := client.GetDepartment("123456")
		if err != nil {
			t.Fatalf("Unexpected error on first call: %v", err)
		}

		// Second call, should get from cache
		_, err = client.GetDepartment("123456")
		if err != nil {
			t.Fatalf("Unexpected error on second call: %v", err)
		}
	})

	t.Run("invalid employee number", func(t *testing.T) {
		_, err := client.GetDepartment("invalid")
		if err == nil {
			t.Error("Expected error for invalid employee number, got nil")
		}
	})

	t.Run("API error response", func(t *testing.T) {
		// Create test server that returns error response
		errorTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `{
				"code": 500,
				"data": null,
				"message": "internal server error",
				"success": false
			}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer errorTS.Close()

		errorClient := NewDepartmentClient(errorTS.URL + "/")
		_, err := errorClient.GetDepartment("123456")
		if err == nil {
			t.Error("Expected error for API error response, got nil")
		}
	})

	t.Run("HTTP error", func(t *testing.T) {
		// Use invalid URL to force HTTP error
		errorClient := NewDepartmentClient("http://invalid-url/")
		errorClient.SetTimeout(time.Millisecond * 100) // Set very short timeout
		_, err := errorClient.GetDepartment("123456")
		if err == nil {
			t.Error("Expected error for HTTP request failure, got nil")
		}
	})
}

func TestCacheExpiration(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"code": 200,
			"data": {
				"dept_1": "a-dept",
				"dept_2": "b-dept",
				"dept_3": "c-dept",
				"dept_4": "d-dept"
			},
			"message": null,
			"success": true
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer ts.Close()

	// Create client
	client := NewDepartmentClient(ts.URL + "/")

	// First call, populate cache
	_, err := client.GetDepartment("123456")
	if err != nil {
		t.Fatalf("Unexpected error on first call: %v", err)
	}

	// Verify cache has data
	client.mutex.RLock()
	item, exists := client.cache["123456"]
	client.mutex.RUnlock()
	if !exists {
		t.Fatal("Expected item to be in cache")
	}

	// Set cache item's expiration to past time to simulate expiration
	client.mutex.Lock()
	item.expiresAt = time.Now().Add(-time.Minute)
	client.cache["123456"] = item
	client.mutex.Unlock()

	// Call again, should get from API again
	_, err = client.GetDepartment("123456")
	if err != nil {
		t.Fatalf("Unexpected error on expired cache call: %v", err)
	}
}

func TestClearCache(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"code": 200,
			"data": {
				"dept_1": "a-dept",
				"dept_2": "b-dept",
				"dept_3": "c-dept",
				"dept_4": "d-dept"
			},
			"message": null,
			"success": true
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer ts.Close()

	// Create client
	client := NewDepartmentClient(ts.URL + "/")

	// Populate cache
	_, err := client.GetDepartment("123456")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify cache has data
	client.mutex.RLock()
	_, exists := client.cache["123456"]
	client.mutex.RUnlock()
	if !exists {
		t.Fatal("Expected item to be in cache")
	}

	// Clear cache
	client.ClearCache()

	// Verify cache is cleared
	client.mutex.RLock()
	_, exists = client.cache["123456"]
	client.mutex.RUnlock()
	if exists {
		t.Error("Expected cache to be cleared")
	}
}
