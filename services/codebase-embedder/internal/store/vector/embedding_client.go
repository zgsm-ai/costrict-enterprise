package vector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/config"
)

// embeddingRequest represents the request body for the embedding API
type embeddingRequest struct {
	Input          []string `json:"input"`           // The text to embed
	Model          string   `json:"model"`           // The model to use for embedding
	EncodingFormat string   `json:"encoding_format"` // The format of the output embeddings
}

// embeddingResponse represents the response body from the embedding API
type embeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// EmbeddingClient defines the interface for embedding service clients
type EmbeddingClient interface {
	// CreateEmbeddings creates embeddings for the given texts using the specified model
	CreateEmbeddings(ctx context.Context, texts []string, model string) ([][]float64, error)
}

// httpEmbeddingClient implements the EmbeddingClient interface using HTTP requests
type httpEmbeddingClient struct {
	config     config.EmbedderConf
	httpClient *http.Client
}

// NewEmbeddingClient creates a new instance of EmbeddingClient
func NewEmbeddingClient(cfg config.EmbedderConf) EmbeddingClient {
	return &httpEmbeddingClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// CreateEmbeddings implements the EmbeddingClient interface
func (c *httpEmbeddingClient) CreateEmbeddings(ctx context.Context, texts []string, model string) ([][]float64, error) {
	requestBody := &embeddingRequest{
		Input:          texts,
		Model:          model,
		EncodingFormat: "float",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.APIBase, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	var attempt int
	var resp *http.Response
	for attempt = 0; attempt <= c.config.MaxRetries; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err == nil {
			break
		}
		if attempt < c.config.MaxRetries {
			time.Sleep(time.Second * time.Duration(attempt+1))
			continue
		}
		return nil, fmt.Errorf("failed to send embedding request after %d retries: %w", attempt, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody := new(bytes.Buffer)
		_, _ = errorBody.ReadFrom(resp.Body)
		return nil, fmt.Errorf("embedding API returned non-OK status %d: %s, body: %s", resp.StatusCode, resp.Status, errorBody.String())
	}

	var responseBody embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response body: %w", err)
	}

	vectors := make([][]float64, len(texts))
	for _, d := range responseBody.Data {
		vectors[d.Index] = d.Embedding
	}
	return vectors, nil
}
