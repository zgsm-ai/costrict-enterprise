package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/types"
)

// GenericClientInterface Generic client interface
type GenericClientInterface interface {
	// Execute Execute tool request
	Execute(ctx context.Context, params map[string]interface{}) (string, error)
	// CheckReady Check service availability
	CheckReady(ctx context.Context, params map[string]interface{}) (bool, error)
}

// CommonParameterNames Define common parameter name constants
const (
	CommonParamClientID      = "clientId"
	CommonParamCodebasePath  = "codebasePath"
	CommonParamClientVersion = "clientVersion"
	CommonParamAuthorization = "authorization"
)

// GetCommonParameterNames Return all common parameter names (excluding authorization)
func GetCommonParameterNames() []string {
	return []string{
		CommonParamClientID,
		CommonParamCodebasePath,
		CommonParamClientVersion,
	}
}

// GenericToolClient Generic client implementation
type GenericToolClient struct {
	toolConfig      config.GenericToolConfig
	searchClient    *HTTPClient
	readyClient     *HTTPClient
	requestBuilder  *GenericRequestBuilder
	responseHandler *GenericResponseHandler
}

// GenericClientFactory Generic client factory
type GenericClientFactory struct {
	clients map[string]GenericClientInterface
	mutex   sync.RWMutex
}

// NewGenericClientFactory Create new generic client factory
func NewGenericClientFactory() *GenericClientFactory {
	return &GenericClientFactory{
		clients: make(map[string]GenericClientInterface),
	}
}

// CreateClient Create client instance based on tool configuration
func (f *GenericClientFactory) CreateClient(toolConfig config.GenericToolConfig) (GenericClientInterface, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Check if cached client already exists
	if client, exists := f.clients[toolConfig.Name]; exists {
		return client, nil
	}

	// Create new generic client
	client, err := f.createGenericClient(toolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create universal client for tool %s: %w", toolConfig.Name, err)
	}

	// Cache client instance
	f.clients[toolConfig.Name] = client
	return client, nil
}

// createGenericClient Create generic client instance
func (f *GenericClientFactory) createGenericClient(toolConfig config.GenericToolConfig) (*GenericToolClient, error) {
	// Configure HTTP client
	searchConfig := HTTPClientConfig{
		Timeout: 5 * time.Second,
	}
	readyConfig := HTTPClientConfig{
		Timeout: 3 * time.Second,
	}

	// Create HTTP clients
	searchClient := NewHTTPClient(toolConfig.Endpoints.Search, searchConfig)
	readyClient := NewHTTPClient(toolConfig.Endpoints.Ready, readyConfig)

	return &GenericToolClient{
		toolConfig:      toolConfig,
		searchClient:    searchClient,
		readyClient:     readyClient,
		requestBuilder:  &GenericRequestBuilder{toolConfig: toolConfig},
		responseHandler: &GenericResponseHandler{},
	}, nil
}

// ClearCache Clear client cache
func (f *GenericClientFactory) ClearCache() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.clients = make(map[string]GenericClientInterface)
}

// Execute Execute tool request
func (c *GenericToolClient) Execute(ctx context.Context, params map[string]interface{}) (string, error) {
	httpReq := c.requestBuilder.BuildRequest(params)

	resp, err := c.searchClient.DoRequest(ctx, httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	return c.responseHandler.HandleResponse(resp)
}

// CheckReady Check service availability
func (c *GenericToolClient) CheckReady(ctx context.Context, params map[string]interface{}) (bool, error) {
	httpReq := c.requestBuilder.BuildReadyRequest(params)

	resp, err := c.readyClient.DoRequest(ctx, httpReq)
	if err != nil {
		return false, fmt.Errorf("failed to check ready status: %w", err)
	}
	defer resp.Body.Close()

	return c.responseHandler.HandleReadyResponse(resp)
}

// GenericRequestBuilder Generic request builder
type GenericRequestBuilder struct {
	toolConfig config.GenericToolConfig
}

// BuildRequest Build request
func (b *GenericRequestBuilder) BuildRequest(params map[string]interface{}) Request {
	req := Request{
		Method: b.toolConfig.Method,
		Headers: map[string]string{
			types.HeaderClientVersion: getStringParam(params, "clientVersion"),
		},
	}

	// Handle parameters based on HTTP method
	if b.toolConfig.Method == http.MethodGet {
		// GET request: build parameters as query string
		req.QueryParams = b.buildQueryParams(params)
	} else {
		// POST request: build parameters as request body
		req.Body = b.buildRequestBody(params)
	}

	// Set authentication information
	if _, exists := params["authorization"]; exists {
		req.Authorization = getStringParam(params, "authorization")
	}

	return req
}

// BuildReadyRequest Build readiness check request
func (b *GenericRequestBuilder) BuildReadyRequest(params map[string]interface{}) Request {
	return Request{
		Method: http.MethodGet,
		Headers: map[string]string{
			types.HeaderClientVersion: getStringParam(params, "clientVersion"),
		},
		QueryParams: map[string]string{
			"clientId":     getStringParam(params, "clientId"),
			"codebasePath": getStringParam(params, "codebasePath"),
		},
		Authorization: getStringParam(params, "authorization"),
	}
}

// getCommonParams Get common parameters (excluding authorization)
func (b *GenericRequestBuilder) getCommonParams(params map[string]interface{}) map[string]interface{} {
	commonParams := make(map[string]interface{})

	commonKeys := []string{CommonParamClientID, CommonParamCodebasePath, CommonParamClientVersion}

	for _, key := range commonKeys {
		if value, exists := params[key]; exists {
			commonParams[key] = value
		}
	}

	return commonParams
}

// buildQueryParams Build query parameters
func (b *GenericRequestBuilder) buildQueryParams(params map[string]interface{}) map[string]string {
	queryParams := make(map[string]string)

	// First add common parameters (if exist), excluding authorization
	commonParams := b.getCommonParams(params)
	for key, value := range commonParams {
		queryParams[key] = fmt.Sprintf("%v", value)
	}

	// Then add parameters according to tool configuration
	for _, param := range b.toolConfig.Parameters {
		if value, exists := params[param.Name]; exists {
			queryParams[param.Name] = fmt.Sprintf("%v", value)
		}
	}

	return queryParams
}

// buildRequestBody Build request body
func (b *GenericRequestBuilder) buildRequestBody(params map[string]interface{}) map[string]interface{} {
	body := make(map[string]interface{})

	// First add common parameters (if exist), excluding authorization
	commonParams := b.getCommonParams(params)
	for key, value := range commonParams {
		body[key] = value
	}

	// Then add parameters according to tool configuration
	for _, param := range b.toolConfig.Parameters {
		if value, exists := params[param.Name]; exists {
			body[param.Name] = value
		}
	}

	return body
}

// GenericResponseHandler Generic response handler
type GenericResponseHandler struct{}

// HandleResponse Handle response
func (h *GenericResponseHandler) HandleResponse(resp *http.Response) (string, error) {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		respBody := ""
		if body != nil {
			respBody = string(body)
		}
		return "", fmt.Errorf(
			"request failed! status: %d, response:%s, url: %s",
			resp.StatusCode, respBody, resp.Request.URL.String(),
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// HandleReadyResponse Handle readiness check response
func (h *GenericResponseHandler) HandleReadyResponse(resp *http.Response) (bool, error) {
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %v", err)
	}

	return false, fmt.Errorf("code: %d, body: %s", resp.StatusCode, body)
}

// getStringParam Get string parameter
func getStringParam(params map[string]interface{}, key string) string {
	if value, exists := params[key]; exists {
		return fmt.Sprintf("%v", value)
	}
	return ""
}
