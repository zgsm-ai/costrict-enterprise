package proxy

import (
	"encoding/json"
	"net/http"
	"time"
)

// ProxyError 代理错误响应
type ProxyError struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorCode 定义错误码常量
const (
	ErrorCodeBadRequest        = "PROXY_BAD_REQUEST"
	ErrorCodeTargetUnreachable = "PROXY_TARGET_UNREACHABLE"
	ErrorCodeTimeout           = "PROXY_TIMEOUT"
	ErrorCodeInternalError     = "PROXY_INTERNAL_ERROR"
)

// CreateProxyError 创建统一错误响应
func CreateProxyError(code, message, details string) *ProxyError {
	return &ProxyError{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().UTC(),
	}
}

// SendErrorResponse 发送错误响应
func SendErrorResponse(w http.ResponseWriter, err *ProxyError, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err != nil {
		json.NewEncoder(w).Encode(err)
	}
}

// NewBadRequestError 创建400错误
func NewBadRequestError(details string) *ProxyError {
	return CreateProxyError(
		ErrorCodeBadRequest,
		"Invalid request format",
		details,
	)
}

// NewTargetUnreachableError 创建503错误
func NewTargetUnreachableError(details string) *ProxyError {
	return CreateProxyError(
		ErrorCodeTargetUnreachable,
		"Target service is unreachable",
		details,
	)
}

// NewTimeoutError 创建504错误
func NewTimeoutError(details string) *ProxyError {
	return CreateProxyError(
		ErrorCodeTimeout,
		"Request timeout",
		details,
	)
}

// NewInternalError 创建500错误
func NewInternalError(details string) *ProxyError {
	return CreateProxyError(
		ErrorCodeInternalError,
		"Internal server error",
		details,
	)
}
