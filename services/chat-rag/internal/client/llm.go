package client

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/timeout"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"github.com/zgsm-ai/chat-rag/internal/utils"
	"go.uber.org/zap"
)

// LLMInterface defines the interface for LLM clients
type LLMInterface interface {
	// GetModelName returns the name of the model
	GetModelName() string
	// GenerateContent directly generates non-streaming content with system prompts and user prompts
	GenerateContent(ctx context.Context, systemPrompt string, userMessages []types.Message) (string, error)
	// ChatLLMWithMessagesStreamRaw directly calls the API using HTTP client to get raw streaming response
	ChatLLMWithMessagesStreamRaw(ctx context.Context, params types.LLMRequestParams, idleTimer *timeout.IdleTimer, callback func(LLMResponse) error) error
	//ChatLLMWithMessagesRaw directly calls the API using HTTP client to get raw non-streaming response
	ChatLLMWithMessagesRaw(ctx context.Context, params types.LLMRequestParams, idleTimer *timeout.IdleTimer) (types.ChatCompletionResponse, error)
	// SetTools sets the tools for the LLM client
	SetTools(tools []types.Function)
}

type LLMResponse struct {
	Header      *http.Header
	ResonseLine string
}

// LLMClient handles communication with language models
type LLMClient struct {
	modelName              string
	endpoint               string
	tools                  []types.Function
	headers                *http.Header
	httpClient             *http.Client
	idleTimeout            time.Duration
	timeoutConfig          config.LLMTimeoutConfig
	StreamChunkInfo        *utils.ChunkStatInfo
	StreamChunkInfoEnabled bool
}

// NewLLMClient creates a new LLM client instance
func NewLLMClient(llmConfig config.LLMConfig, timeoutConfig config.LLMTimeoutConfig, modelName string, headers *http.Header) (LLMInterface, error) {
	// Check for empty endpoint
	if llmConfig.Endpoint == "" || headers == nil {
		return nil, fmt.Errorf("NewLLMClient llmEndpoint cannot be empty")
	}

	idleTimeout := time.Duration(timeoutConfig.IdleTimeoutMs) * time.Millisecond
	if idleTimeout <= 0 {
		idleTimeout = 30 * time.Second
	}

	// Create HTTP client with idle timeout as ResponseHeaderTimeout
	httpClient := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: idleTimeout,
		},
	}

	return &LLMClient{
		modelName:              modelName,
		endpoint:               llmConfig.Endpoint,
		httpClient:             httpClient,
		headers:                headers,
		idleTimeout:            idleTimeout,
		timeoutConfig:          timeoutConfig,
		StreamChunkInfoEnabled: llmConfig.ChunkMetricsEnabled,
	}, nil
}

func (c *LLMClient) GetModelName() string {
	return c.modelName
}

func (c *LLMClient) SetTools(tools []types.Function) {
	c.tools = tools
}

// GenerateContent generate content using a structured message format
func (c *LLMClient) GenerateContent(ctx context.Context, systemPrompt string, userMessages []types.Message) (string, error) {
	// Create a new slice of messages for the summary request
	var messages []types.Message

	// Add system message with the summary prompt
	messages = append(messages, types.Message{
		Role:    types.RoleSystem,
		Content: systemPrompt,
	})

	messages = append(messages, userMessages...)
	// Call ChatLLMWithMessagesRaw to get the raw response
	params := types.LLMRequestParams{
		Messages: messages,
	}

	// Create a simple idle tracker for GenerateContent
	tracker := timeout.NewIdleTracker(time.Duration(c.timeoutConfig.TotalIdleTimeoutMs) * time.Millisecond)
	_, cancel, idleTimer := timeout.NewIdleTimer(ctx, c.idleTimeout, tracker)
	defer func() {
		idleTimer.Stop()
		cancel()
	}()

	result, err := c.ChatLLMWithMessagesRaw(ctx, params, idleTimer)
	if err != nil {
		return "", fmt.Errorf("failed to get response from ChatLLMWithMessagesRaw: %w", err)
	}

	// Check if there are any choices in the response
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	// Extract content from the first choice's message
	content := utils.GetContentAsString(result.Choices[0].Message.Content)
	return content, nil
}

// handleAPIError handles common API error processing for both streaming and non-streaming responses
func (c *LLMClient) handleAPIError(resp *http.Response, logMessage string) error {
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	logger.Warn(logMessage,
		zap.Int("status code", resp.StatusCode),
		zap.String("body", bodyStr),
	)

	// parse err
	var apiError types.APIError
	if err := json.Unmarshal([]byte(bodyStr), &apiError); err == nil {
		apiError.StatusCode = resp.StatusCode
		if strings.Contains(apiError.Code, string(types.ErrQuotaCheck)) {
			apiError.Type = string(types.ErrQuotaCheck)
			return &apiError
		}

		if strings.Contains(apiError.Code, string(types.ErrQuotaManager)) {
			apiError.Type = string(types.ErrQuotaManager)
			return &apiError
		}

		if strings.Contains(apiError.Code, string(types.ErrAiGateway)) {
			apiError.Type = string(types.ErrAiGateway)
			return &apiError
		}
	}

	if bodyStr == "" {
		bodyStr = "None"
	}

	return types.NewHTTPStatusError(resp.StatusCode, bodyStr)
}

// ChatLLMWithMessagesStreamRaw directly calls the API using HTTP client to get raw streaming response
func (c *LLMClient) ChatLLMWithMessagesStreamRaw(ctx context.Context, params types.LLMRequestParams, idleTimer *timeout.IdleTimer, callback func(LLMResponse) error) error {
	if callback == nil {
		return fmt.Errorf("callback function cannot be nil")
	}

	if params.Extra == nil {
		params.Extra = make(map[string]any)
	}
	params.Extra["model"] = c.modelName

	// Prepare request data structure
	requestPayload := types.ChatLLMRequestStream{
		ChatCompletionRequest: types.ChatCompletionRequest{
			Model:            c.modelName,
			LLMRequestParams: params,
		},
		Stream: true,
		StreamOptions: types.StreamOptions{
			IncludeUsage: true,
		},
	}

	// Create request
	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal request payload: %w", err)
	}

	reader := strings.NewReader(string(jsonData))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, reader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set request headers
	for key, values := range *c.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Ensure Content-Length is set correctly
	req.ContentLength = int64(reader.Len())

	// Log before sending request to LLM
	logger.InfoC(ctx, "Starting request to LLM model ...")
	requestStart := time.Now()

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check if it's a timeout error
		if ctx.Err() != nil && idleTimer != nil && idleTimer.IsTimedOut() {
			if idleTimer.Reason() == timeout.IdleTimeoutReasonTotal {
				return types.NewTotalIdleTimeoutError()
			}
			return types.NewStreamIdleTimeoutError()
		}

		// Check if it's a context cancellation (client disconnect)
		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			logger.WarnC(ctx, "Context canceled connecting to LLM service", zap.Error(err))
			return context.Canceled
		}

		logger.ErrorC(ctx, "Failed to connect to LLM service", zap.Error(err))
		return types.NewModelServiceUnavailableError()
	}
	defer resp.Body.Close()

	// Reset idle timer after receiving response headers
	firstByteLatency := time.Since(requestStart)
	if idleTimer != nil {
		idleTimer.Reset()
	}
	logger.InfoC(ctx, "Received response headers from LLM",
		zap.Duration("firstByteLatency", firstByteLatency))

	if resp.StatusCode != http.StatusOK {
		return c.handleAPIError(resp, "LLMClient get straming error response")
	}

	headers := resp.Header
	llmResp := LLMResponse{
		Header: &headers,
	}

	var chunkTimeChan chan float32 = nil
	var chunkTimeCaculated chan bool = nil
	var streamEnd chan bool = nil
	if c.StreamChunkInfoEnabled {
		chunkTimeChan = make(chan float32, 500) // time in ms
		chunkTimeCaculated = make(chan bool, 1) // signal caculate finished
		streamEnd = make(chan bool, 1)          // stream completion
		// steamEnd and chunkTimeChan must close in now go routine
		defer close(streamEnd)
		defer close(chunkTimeChan)
		go func() {
			// calculate chunkTimeCalculated new go routine
			defer close(chunkTimeCaculated) // signal completion
			var stats *utils.ChunkStats = utils.NewChunkStats(0)
			defer stats.Stop()
			for {
				steamIsEnd := false
				select {
				case <-streamEnd:
					steamIsEnd = true
				case chunkTime := <-chunkTimeChan:
					stats.OnChunkArrivedWithInterval(chunkTime)
				case <-ctx.Done(): // reqeust cancel no need to wait
					chunkTimeCaculated <- false
					return
				}
				if steamIsEnd {
					break
				}
			}
			c.StreamChunkInfo = stats.End() // calculate chunk stats
			chunkTimeCaculated <- true      // signal completion
		}()
	}
	// Read streaming response line by line
	scanner := bufio.NewScanner(resp.Body)
	// Increase buffer size to handle long response lines
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	chunkStartTime := time.Now()
	for scanner.Scan() {
		line := scanner.Text()
		if chunkTimeChan != nil {
			// calculate chunk time
			chunkTime := time.Since(chunkStartTime).Microseconds()
			// Non-blocking send to avoid blocking the service if channel is full or closed
			select {
			case chunkTimeChan <- float32(chunkTime) / 1000: // in ms
			default:
				// Skip if channel is full or no receiver
			}
		}

		// Reset idle timer on each line received
		if idleTimer != nil {
			idleTimer.Reset()
		}

		// Arrange non-empty lines, including empty data lines
		if line != "" || strings.HasPrefix(line, "data:") {
			llmResp.ResonseLine = line
			if err := callback(llmResp); err != nil {
				return fmt.Errorf("callback error: %w", err)
			}
		}
		if chunkTimeChan != nil {
			chunkStartTime = time.Now()
		}
	}
	// steam is End
	if streamEnd != nil {
		streamEnd <- true
	}
	if err := scanner.Err(); err != nil {
		// Check if it's a context timeout
		if ctx.Err() != nil && idleTimer != nil && idleTimer.IsTimedOut() {
			if idleTimer.Reason() == timeout.IdleTimeoutReasonTotal {
				return types.NewTotalIdleTimeoutError()
			}
			return types.NewStreamIdleTimeoutError()
		}

		// Check if it's a context cancellation (client disconnect)
		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			logger.WarnC(ctx, "Context canceled reading response", zap.Error(err))
			return context.Canceled
		}

		logger.ErrorC(ctx, "Error reading response", zap.Error(err))
		return types.NewNetWorkError()
	}
	// Wait for chunk time calculation (max 3 seconds)
	if chunkTimeCaculated != nil {
		select {
		case <-chunkTimeCaculated:
			// Calculation completed
		case <-time.After(3 * time.Second): // mas 3 seconds
			// Timeout, proceed without waiting
		case <-ctx.Done():
			// Context cancellation
		}
	}

	return nil
}

// ChatLLMWithMessagesRaw directly calls the API using HTTP client to get raw non-streaming response
func (c *LLMClient) ChatLLMWithMessagesRaw(ctx context.Context, params types.LLMRequestParams, idleTimer *timeout.IdleTimer) (types.ChatCompletionResponse, error) {
	// Prepare request data structure
	if params.Extra == nil {
		params.Extra = make(map[string]any)
	}
	params.Extra["model"] = c.modelName

	requestPayload := types.ChatCompletionRequest{
		Model:            c.modelName,
		LLMRequestParams: params,
	}

	nil_resp := types.ChatCompletionResponse{}

	// Log request data for debugging
	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return nil_resp, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create request
	reader := strings.NewReader(string(jsonData))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, reader)
	if err != nil {
		return nil_resp, fmt.Errorf("failed to create request: %w", err)
	}

	// Set request headers
	for key, values := range *c.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Ensure Content-Length is set correctly
	req.ContentLength = int64(reader.Len())

	requestStart := time.Now()
	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check if it's a timeout error
		if ctx.Err() != nil && idleTimer != nil && idleTimer.IsTimedOut() {
			if idleTimer.Reason() == timeout.IdleTimeoutReasonTotal {
				return nil_resp, types.NewTotalIdleTimeoutError()
			}
			return nil_resp, types.NewStreamIdleTimeoutError()
		}

		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			logger.WarnC(ctx, "Context canceled connecting to LLM service", zap.Error(err))
			return nil_resp, context.Canceled
		}

		logger.ErrorC(ctx, "Failed to connect to LLM service", zap.Error(err))
		return nil_resp, types.NewModelServiceUnavailableError()
	}
	defer resp.Body.Close()

	// Reset idle timer after receiving response headers
	firstByteLatency := time.Since(requestStart)
	if idleTimer != nil {
		idleTimer.Reset()
	}
	logger.InfoC(ctx, "Received response headers from LLM",
		zap.Duration("firstByteLatency", firstByteLatency))

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		err := c.handleAPIError(resp, "LLMClient get error response")
		return nil_resp, err
	}

	// Read response body in chunks, resetting idle timer
	const chunkSize = 8192
	buf := make([]byte, chunkSize)
	var bodyData []byte

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			bodyData = append(bodyData, buf[:n]...)
			// Reset idle timer on each chunk received
			if idleTimer != nil {
				idleTimer.Reset()
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			// Check if it's a timeout error
			if ctx.Err() != nil && idleTimer != nil && idleTimer.IsTimedOut() {
				if idleTimer.Reason() == timeout.IdleTimeoutReasonTotal {
					return nil_resp, types.NewTotalIdleTimeoutError()
				}
				return nil_resp, types.NewStreamIdleTimeoutError()
			}

			if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
				return nil_resp, context.Canceled
			}

			return nil_resp, fmt.Errorf("failed to read response body: %w", err)
		}
	}

	var result types.ChatCompletionResponse
	if err := json.Unmarshal(bodyData, &result); err != nil {
		bodyStr := string(bodyData)
		return nil_resp, fmt.Errorf("failed to parse response (invalid JSON? body: %s)\nerror: %w", bodyStr, err)
	}

	return result, nil
}
