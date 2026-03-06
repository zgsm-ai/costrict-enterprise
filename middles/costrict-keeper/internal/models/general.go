package models

// ErrorResponse defines API error response format
type ErrorResponse struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}
