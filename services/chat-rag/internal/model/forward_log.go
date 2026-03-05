package model

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

// SSEEvent represents a single Server-Sent Event
type SSEEvent struct {
	Event string      `json:"event,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

// ForwardLog represents a forward request log entry
type ForwardLog struct {
	Timestamp time.Time       `json:"timestamp"`
	Request   ForwardRequest  `json:"request"`
	Response  ForwardResponse `json:"response"`
	TargetURL string          `json:"target_url"`
	Duration  time.Duration   `json:"duration_ms"`
	Error     string          `json:"error,omitempty"`
}

// ForwardRequest represents the forwarded request
type ForwardRequest struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   string            `json:"query"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body,omitempty"`
}

// ForwardResponse represents the response from the forwarded request
type ForwardResponse struct {
	StatusCode  int               `json:"status_code"`
	Headers     map[string]string `json:"headers"`
	BodyContent string            `json:"body_content,omitempty"`
	BeautyBody  interface{}       `json:"beauty_body,omitempty"`
	Body        interface{}       `json:"body,omitempty"`
}

// ToJSON converts the forward log to formatted JSON string
func (fl *ForwardLog) ToJSON() (string, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false) // 禁用 HTML 字符转义
	if err := encoder.Encode(fl); err != nil {
		return "", err
	}
	return strings.TrimSpace(buffer.String()), nil
}

// ForwardLogFromJSON creates a ForwardLog from JSON string
func ForwardLogFromJSON(jsonStr string) (*ForwardLog, error) {
	var log ForwardLog
	err := json.Unmarshal([]byte(jsonStr), &log)
	if err != nil {
		return nil, err
	}
	return &log, nil
}
