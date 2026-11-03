package model

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/zgsm-ai/chat-rag/internal/types"
)

// LatencyMetrics represents latency metrics in milliseconds
type LatencyMetrics struct {
	MainModelLatency  int64            `json:"main_model_latency_ms"`
	TotalLatency      int64            `json:"total_latency_ms"`
	FirstTokenLatency int64            `json:"first_token_latency_ms"`
	WindowLatency     int64            `json:"window_latency_ms_ms"`
	ChunkInfo         *StreamChunkInfo `json:"chunk_info"`
}

// ChunkInfo represents chunk interval statistics
type StreamChunkInfo struct {
	ChunkTotal   int     `json:"chunk_total"`
	IntervalAvg  float32 `json:"interval_avg"`
	IntervalMin  float32 `json:"interval_min"`
	IntervalMax  float32 `json:"interval_max"`
	Variance     float32 `json:"variance"`
	StdDeviation float32 `json:"std_deviation"`
	P50          float32 `json:"p50"`
	P95          float32 `json:"p95"`
	P99          float32 `json:"p99"`
}

type ToolCall struct {
	ToolName     string `json:"tool_name"`
	ToolInput    string `json:"tool_input"`
	ToolOutput   string `json:"tool_output"`
	ResultStatus string `json:"result_status"`
	Latency      int64  `json:"latency"`
	Error        string `json:"error"`
}

// RequestParams represents the request parameters for a chat completion
type RequestParams struct {
	Model       string                 `json:"model"`
	RoutedModel string                 `json:"routed_model,omitempty"`
	LlmParams   types.LLMRequestParams `json:"llm_params"`
}

// ChatLog represents a single chat completion log entry
type ChatLog struct {
	WorkspaceInfo WorkspaceInfo `json:"workspace-info"`
	Identity      Identity      `json:"identity"`
	Timestamp     time.Time     `json:"timestamp"`
	// Agent information
	Agent string `json:"agent,omitempty"`
	// Token statistics
	Tokens types.TokenMetrics `json:"tokens"`

	// Processing flags
	IsPromptProceed bool `json:"is_prompt_proceed"`

	// Latency metrics
	Latency LatencyMetrics `json:"latency"`

	// Tools
	ToolCalls []ToolCall `json:"tool_calls"`

	Params RequestParams `json:"params"`

	// OriginalPrompt  []types.Message `json:"original_prompt"`
	ProcessedPrompt []types.Message `json:"processed_prompt"`

	// Response information
	ResponseHeaders []map[string]string    `json:"response_headers,omitempty"`
	ResponseContent *types.ResponseContent `json:"response_content,omitempty"`
	Usage           types.Usage            `json:"usage,omitempty"`

	// Classification (will be filled by async processor)
	Category string `json:"category,omitempty"`

	// Error information
	Error []map[types.ErrorType]string `json:"error,omitempty"`

	// Degradation trace
	DegradationTrace []DegradationEvent `json:"degradation_trace,omitempty"`
}

type DegradationEventType string

const (
	DegradationEventStarted                        DegradationEventType = "degradation_started"
	DegradationEventAttemptStarted                 DegradationEventType = "attempt_started"
	DegradationEventAttemptFailed                  DegradationEventType = "attempt_failed"
	DegradationEventRetryScheduled                 DegradationEventType = "retry_scheduled"
	DegradationEventRetrySkippedInsufficientBudget DegradationEventType = "retry_skipped_insufficient_budget"
	DegradationEventModelSwitch                    DegradationEventType = "model_switch"
	DegradationEventStopped                        DegradationEventType = "degradation_stopped"
	DegradationEventModelSucceeded                 DegradationEventType = "model_succeeded"
	DegradationEventTraceTruncated                 DegradationEventType = "trace_truncated"
)

type DegradationEvent struct {
	Sequence            int                  `json:"sequence"`
	Timestamp           time.Time            `json:"timestamp"`
	Event               DegradationEventType `json:"event"`
	Model               string               `json:"model,omitempty"`
	PreviousModel       string               `json:"previous_model,omitempty"`
	OrderedModels       []string             `json:"ordered_models,omitempty"`
	ModelIndex          int                  `json:"model_index"`
	Attempt             int                  `json:"attempt,omitempty"`
	MaxRetries          int                  `json:"max_retries,omitempty"`
	Retryable           *bool                `json:"retryable,omitempty"`
	Switchable          *bool                `json:"switchable,omitempty"`
	StatusCode          int                  `json:"status_code,omitempty"`
	ErrorCode           string               `json:"error_code,omitempty"`
	ErrorType           string               `json:"error_type,omitempty"`
	ErrorMessage        string               `json:"error_message,omitempty"`
	ElapsedMs           int64                `json:"elapsed_ms,omitempty"`
	Reason              string               `json:"reason,omitempty"`
	CompatibilitySignal bool                 `json:"compatibility_signal,omitempty"`
}

// toStringJSON converts the log entry to indented JSON string
func (cl *ChatLog) toStringJSON(indent string) (string, error) {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", indent)
	err := encoder.Encode(cl)
	if err != nil {
		return "", err
	}
	// Remove the newline added by Encode()
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

// ToCompressedJSON converts the log entry to JSON string
func (cl *ChatLog) ToCompressedJSON() (string, error) {
	return cl.toStringJSON("")
}

// ToPrettyJSON Using 2 spaces for compact yet readable indentation (standard JSON formatting practice)
func (cl *ChatLog) ToPrettyJSON() (string, error) {
	return cl.toStringJSON("  ")
}

// FromJSON creates a ChatLog from JSON string
func FromJSON(jsonStr string) (*ChatLog, error) {
	var log ChatLog
	err := json.Unmarshal([]byte(jsonStr), &log)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// Helper functions

// AddError adds an error entry with type and message to the ChatLog
func (cl *ChatLog) AddError(errorType types.ErrorType, err error) {
	if cl.Error == nil {
		cl.Error = make([]map[types.ErrorType]string, 0)
	}
	cl.Error = append(cl.Error, map[types.ErrorType]string{
		errorType: err.Error(),
	})
}
