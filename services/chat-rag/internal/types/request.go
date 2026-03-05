package types

import (
	"bytes"
	"encoding/json"
)

const (
	// RoleSystem System role message
	RoleSystem = "system"

	// RoleUser User role message
	RoleUser = "user"

	// RoleTool Tool role message
	RoleTool = "tool"

	// RoleAssistant AI assistant role message
	RoleAssistant = "assistant"
)

// PromptMode defines different types of chat
type PromptMode string

const (
	// Raw mode: No deep processing of user prompt, only necessary operations like compression
	Raw PromptMode = "raw"

	// Balanced mode: Considering both cost and performance, choosing a compromise approach
	// including rag and prompt compression
	Balanced PromptMode = "balanced"

	// Cost Cost-first mode: Minimizing LLM calls and context size to save cost
	Cost PromptMode = "cost"

	// Performance Performance-first mode: Maximizing output quality without considering cost
	Performance PromptMode = "performance"

	// Auto select mode: Default is balanced mode
	Auto PromptMode = "auto"

	// Strict mode: Strictly follow the workflow agent
	Strict PromptMode = "strict"
)

const (
	// Request Headers
	HeaderQuotaIdentity = "x-quota-identity"
	HeaderRequestId     = "x-request-id"
	HeaderCaller        = "x-caller"
	HeaderTaskId        = "zgsm-task-id"
	HeaderClientId      = "zgsm-client-id"
	HeaderClientIde     = "zgsm-client-ide"
	HeaderClientOS      = "X-Stainless-OS"
	HeaderLanguage      = "Accept-Language"
	HeaderAuthorization = "authorization"
	HeaderProjectPath   = "zgsm-project-path"
	HeaderClientVersion = "X-Costrict-Version"
	HeaderOriginalModel = "x-original-model"

	// Response Headers
	HeaderUserInput   = "x-user-input"
	HeaderSelectLLm   = "x-select-llm"
	HeaderOneAPIReqId = "x-oneapi-request-id"
)

// ResponseHeadersToForward defines the list of response headers that should be forwarded
var ResponseHeadersToForward = []string{
	HeaderUserInput,
	HeaderSelectLLm,
	HeaderOneAPIReqId,
}

// ToolStatus defines the status of the tool
type ToolStatus string

const (
	ToolStatusRunning ToolStatus = "running"
	ToolStatusSuccess ToolStatus = "success"
	ToolStatusFailed  ToolStatus = "failed"
)

// Redis key prefix for tool status
const ToolStatusRedisKeyPrefix = "tool_status:"

// Tool string filter
const StrFilterToolAnalyzing = "\n#### 💡 检索已完成，分析中"
const StrFilterToolSearchStart = "\n#### 🔍 "
const StrFilterToolSearchEnd = "工具检索中"

type ExtraBody struct {
	PromptMode PromptMode `json:"prompt_mode,omitempty"`
	Mode       string     `json:"mode,omitempty"`

	// Extra fields for transparent passthrough of unknown fields
	Extra map[string]any `json:"-"`
}

// UnmarshalJSON implements custom JSON unmarshaling to capture unknown fields
func (e *ExtraBody) UnmarshalJSON(data []byte) error {
	// First unmarshal into a map to capture all fields
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Extract known fields
	if promptMode, ok := raw["prompt_mode"].(string); ok {
		e.PromptMode = PromptMode(promptMode)
		delete(raw, "prompt_mode")
	}
	if mode, ok := raw["mode"].(string); ok {
		e.Mode = mode
		delete(raw, "mode")
	}

	// Store remaining fields in Extra for passthrough
	if len(raw) > 0 {
		e.Extra = raw
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling to include Extra fields
func (e ExtraBody) MarshalJSON() ([]byte, error) {
	// Start with a map containing known fields
	result := make(map[string]any)

	if e.PromptMode != "" {
		result["prompt_mode"] = e.PromptMode
	}
	if e.Mode != "" {
		result["mode"] = e.Mode
	}

	// Merge Extra fields
	for k, v := range e.Extra {
		result[k] = v
	}

	return marshalJSONWithoutEscape(result)
}

type ChatCompletionResponse struct {
	Id      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// LLMRequestParams contains parameters for LLM requests
type LLMRequestParams struct {
	Priority  *int      `json:"priority,omitempty"`
	ExtraBody ExtraBody `json:"extra_body,omitempty"`
	Messages  []Message `json:"messages"`

	// Extra fields for transparent passthrough of unknown fields like tools, functions, max_tokens, temperature, etc.
	Extra map[string]any `json:"-"`
}

// UnmarshalJSON implements custom JSON unmarshaling to capture unknown fields
func (p *LLMRequestParams) UnmarshalJSON(data []byte) error {
	// First unmarshal into a map to capture all fields
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Extract known fields
	if priority, ok := raw["priority"]; ok {
		if priorityFloat, ok := priority.(float64); ok {
			priorityInt := int(priorityFloat)
			p.Priority = &priorityInt
		}
		delete(raw, "priority")
	}
	if extraBody, ok := raw["extra_body"]; ok {
		extraBodyBytes, _ := json.Marshal(extraBody)
		json.Unmarshal(extraBodyBytes, &p.ExtraBody)
		delete(raw, "extra_body")
	}
	if messages, ok := raw["messages"]; ok {
		messagesBytes, _ := json.Marshal(messages)
		json.Unmarshal(messagesBytes, &p.Messages)
		delete(raw, "messages")
	}

	// Store remaining fields in Extra for passthrough
	if len(raw) > 0 {
		p.Extra = raw
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling to include Extra fields
func (p LLMRequestParams) MarshalJSON() ([]byte, error) {
	// Start with a map containing known fields
	result := make(map[string]any)

	if p.Priority != nil {
		result["priority"] = p.Priority
	}
	// Only add extra_body if it has content
	if p.ExtraBody.PromptMode != "" || p.ExtraBody.Mode != "" || len(p.ExtraBody.Extra) > 0 {
		result["extra_body"] = p.ExtraBody
	}
	if p.Messages != nil {
		result["messages"] = p.Messages
	}

	// Merge Extra fields
	for k, v := range p.Extra {
		result[k] = v
	}

	return marshalJSONWithoutEscape(result)
}

type ChatCompletionRequest struct {
	Model            string `json:"model"`
	LLMRequestParams        // Embedded params
}

// UnmarshalJSON implements custom JSON unmarshaling for ChatCompletionRequest
// This is needed because the embedded LLMRequestParams has custom unmarshaling that captures unknown fields
func (r *ChatCompletionRequest) UnmarshalJSON(data []byte) error {
	// First unmarshal into a map to capture all fields
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Extract model field before passing to embedded type
	if model, ok := raw["model"].(string); ok {
		r.Model = model
		delete(raw, "model")
	}

	// Marshal remaining fields and unmarshal into LLMRequestParams
	remainingData, _ := json.Marshal(raw)
	return json.Unmarshal(remainingData, &r.LLMRequestParams)
}

type ChatLLMRequestStream struct {
	ChatCompletionRequest               // Embedded ChatLLMRequest
	Stream                bool          `json:"stream,omitempty"`
	StreamOptions         StreamOptions `json:"stream_options,omitempty"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message,omitempty"`
	Delta        Delta   `json:"delta,omitempty"`
	FinishReason string  `json:"finish_reason,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`

	// Extra fields for transparent passthrough of unknown fields like tool_calls, tool_call_id, name, etc.
	Extra map[string]any `json:"-"`
}

// UnmarshalJSON implements custom JSON unmarshaling to capture unknown fields
func (m *Message) UnmarshalJSON(data []byte) error {
	// First unmarshal into a map to capture all fields
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Extract known fields
	if role, ok := raw["role"].(string); ok {
		m.Role = role
		delete(raw, "role")
	}
	if content, ok := raw["content"]; ok {
		m.Content = content
		delete(raw, "content")
	}

	// Store remaining fields in Extra for passthrough
	if len(raw) > 0 {
		m.Extra = raw
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling to include Extra fields
func (m Message) MarshalJSON() ([]byte, error) {
	// Start with a map containing known fields
	result := make(map[string]any)

	result["role"] = m.Role
	result["content"] = m.Content

	// Merge Extra fields
	for k, v := range m.Extra {
		result[k] = v
	}

	return marshalJSONWithoutEscape(result)
}

type Delta struct {
	Role             string `json:"role,omitempty"`
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
	ToolCalls        []any  `json:"tool_calls,omitempty"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// FunctionCall is the structure of the function called by the LLM.
type Function struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

type FunctionDefinition struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Parameters  FunctionParameters `json:"parameters"`
}

type FunctionParameters struct {
	Type       string                     `json:"type"`
	Properties map[string]PropertyDetails `json:"properties"`
	Required   []string                   `json:"required"`
}

type PropertyDetails struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Items       *Items      `json:"items,omitempty"` // 用于array类型
}

type Items struct {
	Type string `json:"type"`
}

// ToolStatusResponse defines tool status response structure
type ToolStatusResponse struct {
	Code    int            `json:"code"`
	Data    ToolStatusData `json:"data"`
	Message string         `json:"message"`
}

// ToolStatusData defines tool status data structure
type ToolStatusData struct {
	Tools map[string]ToolStatusDetail `json:"tools,omitempty"`
}

// ToolStatusDetail defines tool status detail structure
type ToolStatusDetail struct {
	Status string      `json:"status"`
	Result interface{} `json:"result,omitempty"`
}

// marshalJSONWithoutEscape marshals JSON without HTML escaping
func marshalJSONWithoutEscape(v any) ([]byte, error) {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	// Remove the trailing newline added by Encode
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
