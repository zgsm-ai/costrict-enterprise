package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/zgsm-ai/chat-rag/internal/model"
)

type DepartmentInterface interface {
	// GetDepartment queries department info by employee number
	GetDepartment(employeeNumber string) (*model.DepartmentInfo, error)
}

// Cache item structure
type cacheItem struct {
	data      model.DepartmentInfo
	expiresAt time.Time
}

// DepartmentClient department query client
type DepartmentClient struct {
	url     string
	cache   map[string]cacheItem
	mutex   sync.RWMutex
	timeout time.Duration
}

// Cache valid for 1 week
const cacheExpiration = time.Hour * 24 * 7

// NewDepartmentClient creates a new department query client
func NewDepartmentClient(url string) *DepartmentClient {
	return &DepartmentClient{
		url:     url,
		cache:   make(map[string]cacheItem),
		timeout: time.Second * 10, // Default HTTP request timeout
	}
}

// SetTimeout sets HTTP request timeout
func (c *DepartmentClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetDepartment queries department info by employee number
func (c *DepartmentClient) GetDepartment(employeeNumber string) (*model.DepartmentInfo, error) {
	// First try to get from cache
	if item, ok := c.getFromCache(employeeNumber); ok {
		return &item.data, nil
	}

	// If cache doesn't exist or expired, call API
	url := c.url + employeeNumber

	client := &http.Client{Timeout: c.timeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call department API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("department API returned non-200 status: %d, failed to read body: %v", resp.StatusCode, err)
		}
		defer resp.Body.Close()
		return nil, fmt.Errorf("department API returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var result struct {
		Code    int                  `json:"code"`
		Data    model.DepartmentInfo `json:"data"`
		Message string               `json:"message"`
		Success bool                 `json:"success"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if !result.Success || result.Code != 200 {
		return nil, fmt.Errorf("API returned error: %s", result.Message)
	}

	c.setToCache(employeeNumber, result.Data, cacheExpiration)

	return &result.Data, nil
}

// Get data from cache
func (c *DepartmentClient) getFromCache(employeeNumber string) (cacheItem, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.cache[employeeNumber]
	if !exists {
		return cacheItem{}, false
	}

	if time.Now().After(item.expiresAt) {
		return cacheItem{}, false
	}

	return item, true
}

// Set data to cache
func (c *DepartmentClient) setToCache(employeeNumber string, data model.DepartmentInfo, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[employeeNumber] = cacheItem{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}
}

// ClearCache clears all cached data
func (c *DepartmentClient) ClearCache() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]cacheItem)
}
