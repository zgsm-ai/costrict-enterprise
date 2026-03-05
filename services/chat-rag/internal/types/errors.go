package types

import (
	"fmt"
	"net/http"
)

// ErrorType defines different types of errors
type ErrorType string

const (
	// ErrSemantic represents semantic processing errors
	ErrSemantic ErrorType = "SemanticError"

	// ErrSummary represents summary generation errors
	ErrSummary ErrorType = "SummaryError"

	// ErrApi represents dependent API errors
	ErrApiError ErrorType = "ApiError"

	// ErrServer represents internal server errors
	ErrServerError ErrorType = "ServerError"

	// ErrServer represents context length exceeded
	ErrContextExceeded ErrorType = "ContextLengthExceeded"

	// ErrExtra represents extra operation errors
	ErrExtra ErrorType = "ExtraError"

	// llm api error type
	ErrQuotaCheck   ErrorType = "quota-check"
	ErrQuotaManager ErrorType = "quota-manager"
	ErrAiGateway    ErrorType = "ai-gateway"

	ErrServerModel ErrorType = "ai_model_error"
)

const (
	ErrCodeContextExceeded = "chat-rag.context_length_exceeded"
	ErrMsgContextExceeded  = "The request exceeds the model's maximum context length. Please reduce the length of your input."

	ErrCodeModelServiceUnavailable = "chat-rag.model_services_unavailable"
	ErrMsgModelServiceUnavailable  = "Unable to access the AI model services. Please try again later."

	ErrCodeModelUnavailable = "chat-rag.model_unavailable"
	ErrMsgModelUnavailable  = "The current model is not available, please try another one."

	ErrCodeUnauthorized = "chat-rag.model_services_unauthorized"
	ErrMsgUnauthorized  = "Unauthorized access to model services. Please check your permissions."

	ErrCodeTooManyRequests = "chat-rag.too_many_requests"
	ErrMsgTooManyRequests  = "The number of requests to the model or the number of tokens has reached the official limit."

	ErrCodeInernalError = "chat-rag.internal_error"
	ErrMsgInernalError  = "Internal Server Error. Please try again later."

	ErrCodeNetworkError = "chat-rag.network_interrupt"
	ErrMsgNetworkError  = "Network connection has been interrupted. Please verify your network connectivity and retry."

	ErrCodeServerBusy = "chat-rag.server_busy"
	ErrMsgServerBusy  = "Server is busy. Please try again later."

	ErrCodeStreamIdleTimeout      = "chat-rag.stream_idle_timeout"
	ErrMsgStreamIdleTimeout       = "Request idle timeout: no data received within the allowed idle period."
	ErrCodeTotalStreamIdleTimeout = "chat-rag.total_stream_idle_timeout"
	ErrMsgTotalStreamIdleTimeout  = "Total idle timeout: cumulative idle time across retries exceeded the allowed limit."

	ErrCodeInvalidResponseContent = "chat-rag.invalid_response_content"
	ErrMsgInvalidResponseContent  = "The model is unable to perform inference or makes errors during inference."
)

type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Success    bool   `json:"success"`
	StatusCode int    `json:"statusCode,omitempty"`
	Type       string `json:"type,omitempty"`
}

func NewContextTooLongError() *APIError {
	return &APIError{
		Code:       ErrCodeContextExceeded,
		Message:    ErrMsgContextExceeded,
		Success:    false,
		StatusCode: http.StatusBadRequest,
		Type:       string(ErrServerModel),
	}
}

func NewModelServiceUnavailableError() *APIError {
	return &APIError{
		Code:       ErrCodeModelServiceUnavailable,
		Message:    ErrMsgModelServiceUnavailable,
		Success:    false,
		StatusCode: http.StatusServiceUnavailable,
		Type:       string(ErrServerModel),
	}
}

func NewNetWorkError() *APIError {
	return &APIError{
		Code:       ErrCodeNetworkError,
		Message:    ErrMsgNetworkError,
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Type:       string(ErrServerModel),
	}
}

func NewHTTPStatusError(statusCode int, bodyStr string) *APIError {
	var code string
	var msg string

	switch statusCode {
	case http.StatusNotFound:
		code = ErrCodeModelUnavailable
		msg = ErrMsgModelUnavailable
	case http.StatusUnauthorized:
		code = ErrCodeUnauthorized
		msg = ErrMsgUnauthorized
	case http.StatusTooManyRequests:
		code = ErrCodeTooManyRequests
		msg = ErrMsgTooManyRequests
	case http.StatusRequestEntityTooLarge:
		code = ErrCodeContextExceeded
		msg = ErrMsgContextExceeded
	case http.StatusGatewayTimeout:
		code = ErrCodeServerBusy
		msg = ErrMsgServerBusy
	default:
		code = ErrCodeModelServiceUnavailable
		msg = fmt.Sprintf("%s\n\n[Error Detail]:\nCode: %d\nMessage: %s",
			ErrMsgModelServiceUnavailable, statusCode, bodyStr)
	}

	return &APIError{
		Code:       code,
		Message:    msg,
		Success:    false,
		StatusCode: statusCode,
		Type:       string(ErrServerModel),
	}
}

func NewInvaildResponseContentError() *APIError {
	return &APIError{
		Code:       ErrCodeInvalidResponseContent,
		Message:    ErrMsgInvalidResponseContent,
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Type:       string(ErrServerModel),
	}
}

func (e *APIError) Error() string {
	return fmt.Sprintf(`{"code":"%s","message":"%s","success":%v}`, e.Code, e.Message, e.Success)
}

// IdleTimeoutError represents an idle timeout error
type IdleTimeoutError struct {
	Total      bool   // true if total idle budget exhausted
	StatusCode int    // HTTP status code (504)
	Code       string // Error code
	Message    string // Error message
}

func NewStreamIdleTimeoutError() *IdleTimeoutError {
	return &IdleTimeoutError{
		Total:      false,
		StatusCode: http.StatusOK,
		Code:       ErrCodeStreamIdleTimeout,
		Message:    ErrMsgStreamIdleTimeout,
	}
}

func NewTotalIdleTimeoutError() *IdleTimeoutError {
	return &IdleTimeoutError{
		Total:      true,
		StatusCode: http.StatusOK,
		Code:       ErrCodeTotalStreamIdleTimeout,
		Message:    ErrMsgTotalStreamIdleTimeout,
	}
}

func (e *IdleTimeoutError) Error() string {
	timeoutType := "single"
	if e.Total {
		timeoutType = "total"
	}
	return fmt.Sprintf("idle timeout (%s): %s", timeoutType, e.Message)
}
