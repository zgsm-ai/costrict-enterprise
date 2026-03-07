package response

import "fmt"

type codeMsg struct {
	Code    int
	Message string
}

func (c *codeMsg) Error() string {
	return fmt.Sprintf("code: %d, message: %s", c.Code, c.Message)
}

// NewError creates a new codeMsg.
func NewError(code int, msg string) error {
	return &codeMsg{Code: code, Message: msg}
}

// NewParamError creates a new parameter error.
func NewParamError(msg string) error {
	return &codeMsg{Code: 400, Message: msg}
}

// NewAuthError creates a new authentication error.
func NewAuthError(msg string) error {
	return &codeMsg{Code: 401, Message: msg}
}

// NewPermissionError creates a new permission error.
func NewPermissionError(msg string) error {
	return &codeMsg{Code: 403, Message: msg}
}

// NewRateLimitError creates a new rate limit error.
func NewRateLimitError(msg string) error {
	return &codeMsg{Code: 429, Message: msg}
}
