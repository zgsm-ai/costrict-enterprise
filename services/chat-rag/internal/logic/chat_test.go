package logic

import (
	"context"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/client"
	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/service/mocks"
	"github.com/zgsm-ai/chat-rag/internal/tokenizer"
	"github.com/zgsm-ai/chat-rag/internal/types"
)

// createTestContext creates a test context
func createTestContext() context.Context {
	return context.Background()
}

// createTestServiceContext creates a test ServiceContext
// tokenCounter parameter can be nil or *utils.TokenCounter type
func createTestServiceMock(t *testing.T) (*gomock.Controller, *mocks.MockLoggerInterface, *mocks.MockMetricsInterface) {
	ctrl := gomock.NewController(t)
	loggerMock := mocks.NewMockLoggerInterface(ctrl)
	metricsMock := mocks.NewMockMetricsInterface(ctrl)

	// Setup common mock expectations
	loggerMock.EXPECT().LogAsync(gomock.Any(), gomock.Any()).AnyTimes()
	loggerMock.EXPECT().SetMetricsService(gomock.Any()).AnyTimes()
	metricsMock.EXPECT().GetRegistry().Return(prometheus.NewRegistry()).AnyTimes()
	metricsMock.EXPECT().RecordChatLog(gomock.Any()).AnyTimes()

	return ctrl, loggerMock, metricsMock
}

func createTestServiceContext(t *testing.T, cfg *config.Config, tokenCounter interface{}) *bootstrap.ServiceContext {
	ctrl, loggerMock, metricsMock := createTestServiceMock(t)
	defer ctrl.Finish()

	svcCtx := &bootstrap.ServiceContext{
		Config: config.Config{
			LLM: cfg.LLM,
		},
		LoggerService:  loggerMock,
		MetricsService: metricsMock,
	}

	// If tokenCounter exists and type is correct, set it to ServiceContext
	if tc, ok := tokenCounter.(*tokenizer.TokenCounter); ok {
		svcCtx.TokenCounter = tc
	}

	return svcCtx
}

// createTestRequest creates a test ChatCompletionRequest
func createTestRequest(model string, messages []types.Message, stream bool) *types.ChatCompletionRequest {
	req := &types.ChatCompletionRequest{
		Model: model,
		LLMRequestParams: types.LLMRequestParams{
			Messages: messages,
		},
	}
	// Add stream to Extra map
	if req.Extra == nil {
		req.Extra = make(map[string]any)
	}
	req.Extra["stream"] = stream
	return req
}

// createTestIdentity creates a test Identity
func createTestIdentity() *model.Identity {
	return &model.Identity{
		ClientID:    "test-client",
		ProjectPath: "/test/path",
	}
}

// setupTestLogic combines all helper functions to create complete test logic
func setupTestLogic(t *testing.T, cfg *config.Config, tokenCounter interface{},
	model string, messages []types.Message, writer http.ResponseWriter) (*ChatCompletionLogic, *bootstrap.ServiceContext) {
	ctx := createTestContext()
	svcCtx := createTestServiceContext(t, cfg, tokenCounter)
	req := createTestRequest(model, messages, false)
	identity := createTestIdentity()
	headers := make(http.Header)

	// Set mock expectations
	if logger, ok := svcCtx.LoggerService.(*mocks.MockLoggerInterface); ok {
		logger.EXPECT().LogAsync(gomock.Any(), gomock.Any()).AnyTimes()
	}

	return NewChatCompletionLogic(ctx, svcCtx, req, writer, &headers, identity), svcCtx
}

func TestChatCompletionLogic_NewChatCompletionLogic(t *testing.T) {
	mockWriter := &mockResponseWriter{}
	cfg := &config.Config{}
	logic, svcCtx := setupTestLogic(t, cfg, nil, "test-model", []types.Message{
		{Role: "user", Content: "Hello"},
	}, mockWriter)

	assert.NotNil(t, logic)
	assert.Equal(t, createTestContext(), logic.ctx)
	assert.Equal(t, svcCtx, logic.svcCtx)
}

// TestLLMClientMock tests LLMClient mock
func TestLLMClientMock(t *testing.T) {
	mock := &client.LLMClient{}
	assert.NotNil(t, mock)
}

func TestChatCompletionLogic_countTokensInMessages_Fallback(t *testing.T) {
	cfg := &config.Config{}
	logic, _ := setupTestLogic(t, cfg, nil, "test-model", []types.Message{}, &mockResponseWriter{})

	messages := []types.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}

	count := logic.countTokensInMessages(messages)
	assert.Greater(t, count, 0) // Should return estimated token count
}

func TestChatCompletionLogic_ChatCompletion_ValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.Config
		expected string
	}{
		{
			name:     "empty endpoint",
			config:   &config.Config{},
			expected: "NewLLMClient llmEndpoint cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logic, _ := setupTestLogic(t, tt.config, nil, "test-model", []types.Message{}, &mockResponseWriter{})

			resp, err := logic.ChatCompletion()
			t.Log("==>", err)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expected)
			assert.Nil(t, resp)
		})
	}

	// Test valid configuration
	t.Run("valid config", func(t *testing.T) {
		cfg := &config.Config{
			LLM: config.LLMConfig{
				Endpoint: "http://test-endpoint",
			},
		}
		_, svcCtx := setupTestLogic(t, cfg, nil, "test-model", []types.Message{}, &mockResponseWriter{})
		assert.NotNil(t, svcCtx)
		assert.Equal(t, "http://test-endpoint", svcCtx.Config.LLM.Endpoint)
	})
}

// Tests TokenCounter setup correctly
func TestChatCompletionLogic_WithTokenCounter(t *testing.T) {
	mockWriter := &mockResponseWriter{}
	cfg := &config.Config{}
	tokenCounter := &tokenizer.TokenCounter{}

	logic, svcCtx := setupTestLogic(t, cfg, tokenCounter, "test-model",
		[]types.Message{{Role: "user", Content: "Hello"}}, mockWriter)

	assert.NotNil(t, logic)
	assert.NotNil(t, svcCtx.TokenCounter)
	assert.Equal(t, tokenCounter, svcCtx.TokenCounter)
}

func TestChatCompletionLogic_ChatCompletion_BasicRequest(t *testing.T) {
	cfg := config.MustLoadConfig("../../etc/chat-api.yaml")

	// Initialize token counter
	tokenCounter, err := tokenizer.NewTokenCounter()
	if err != nil {
		tokenCounter = &tokenizer.TokenCounter{} // Fallback to basic counter
	}

	ctrl, _, _ := createTestServiceMock(t)
	defer ctrl.Finish()

	logic, _ := setupTestLogic(t, &cfg, tokenCounter,
		"gpt-3.5-turbo", []types.Message{
			{Role: "user", Content: "Hello, how are you?"},
		}, &mockResponseWriter{})

	// Test basic request
	resp, err := logic.ChatCompletion()

	// Verify response
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// mockResponseWriter mocks http.ResponseWriter and http.Flusher for testing
type mockResponseWriter struct {
	data       []byte
	headers    http.Header
	statusCode int
	flushed    bool
}

func (m *mockResponseWriter) Header() http.Header {
	if m.headers == nil {
		m.headers = make(http.Header)
	}
	return m.headers
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	m.data = append(m.data, data...)
	return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

// Flush implements http.Flusher interface
func (m *mockResponseWriter) Flush() {
	m.flushed = true
}

func TestChatCompletionLogic_ChatCompletion_StreamingRequest(t *testing.T) {
	// Load config
	cfg := config.MustLoadConfig("../../etc/chat-api.yaml")

	// Initialize token counter
	tokenCounter, _ := tokenizer.NewTokenCounter()

	// Setup mocks
	ctrl, loggerMock, metricsMock := createTestServiceMock(t)
	defer ctrl.Finish()

	// Prepare test data
	testModel := "gpt-3.5-turbo"
	testMessages := []types.Message{
		{Role: "user", Content: "Hello, how are you?"},
	}
	testWriter := &mockResponseWriter{}

	// Create service context
	svcCtx := &bootstrap.ServiceContext{
		Config:         cfg,
		LoggerService:  loggerMock,
		MetricsService: metricsMock,
		TokenCounter:   tokenCounter,
	}

	headers := make(http.Header)

	// Create logic instance
	logic := NewChatCompletionLogic(
		createTestContext(),
		svcCtx,
		createTestRequest(testModel, testMessages, true),
		testWriter,
		&headers,
		createTestIdentity(),
	)

	// Execute test
	err := logic.ChatCompletionStream()

	// Verify expected error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401 Authorization Required", "Expected 401 unauthorized error")

	// Verify response write attempt
	assert.Greater(t, len(testWriter.data), 0, "Expected response attempt data")
	assert.True(t, testWriter.flushed, "Expected response flush attempt")
}
