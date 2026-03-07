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
