package helper

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"go.uber.org/zap"
)

// SetSSEResponseHeaders sets SSE response headers
func SetSSEResponseHeaders(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")
	c.Header("X-Accel-Buffering", "no")
}

// SendErrorResponse sends a structured error response
func SendErrorResponse(c *gin.Context, statusCode int, err error) {
	fmt.Printf("==> sendErrorResponse: %+v\n", err)
	message := err.Error()
	errType := "server_error"

	// Check if the error is an APIError with a specific status code
	if apiErr, ok := err.(*types.APIError); ok {
		statusCode = apiErr.StatusCode
		message = apiErr.Message
		errType = apiErr.Type
	}

	c.JSON(statusCode, gin.H{
		"error": map[string]interface{}{
			"message": message,
			"type":    errType,
		},
	})
}

// SendSSEResponseMessage sends a message using SSE format with template rendering
func SendSSEResponseMessage(c *gin.Context, clientIDE string, templateString string, templateData map[string]interface{}) {
	SetSSEResponseHeaders(c)
	c.Status(http.StatusOK)
	logger.InfoC(c, "sending sse response message", zap.String("client_ide", clientIDE))

	const vscode = "Visual Studio Code"
	// Parse and execute template
	if clientIDE == vscode {
		templateString = fmt.Sprintf("{\"result\": \"%s\"}",
			strings.ReplaceAll(templateString, "\n", "\\n"))
	}
	tmpl, err := template.New("sse").Parse(templateString)

	var responseData string
	if err != nil {
		logger.Error("Failed to parse SSE template", zap.Error(err))
		responseData = templateString
	} else {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, templateData); err != nil {
			logger.Error("Failed to execute SSE template", zap.Error(err))
			responseData = templateString
		} else {
			responseData = buf.String()
		}
	}

	generateRandomID := func() string {
		b := make([]byte, 16)
		rand.Read(b)
		return hex.EncodeToString(b)
	}

	randomID := generateRandomID()

	var response interface{}
	if clientIDE == vscode {
		response = types.ChatCompletionResponse{
			Id:      randomID,
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   "",
			Choices: []types.Choice{
				{
					Index: 0,
					Delta: types.Delta{
						Role:             "assistant",
						ReasoningContent: "",
						ToolCalls: []any{
							map[string]interface{}{
								"index": 0,
								"id":    randomID,
								"type":  "function",
								"function": map[string]interface{}{
									"name":      "attempt_completion",
									"arguments": responseData,
								},
							},
						},
					},
				},
			},
		}
	} else {
		response = map[string]interface{}{
			"id":      randomID,
			"object":  "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model":   "",
			"choices": []interface{}{
				map[string]interface{}{
					"index": 0,
					"delta": map[string]interface{}{
						"role":              "assistant",
						"content":           responseData,
						"reasoning_content": "",
						"tool_calls":        nil,
					},
					"logprobs":      nil,
					"finish_reason": "stop",
				},
			},
			"usage": nil,
		}
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal ChatCompletionResponse", zap.Error(err))
		_, err = fmt.Fprintf(c.Writer, "data: %s\n\n", responseData)
	} else {
		_, err = fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
	}
	if err != nil {
		logger.Error("Failed to write SSE response", zap.Error(err))
	}

	flusher, ok := c.Writer.(http.Flusher)
	if ok {
		flusher.Flush()
	}

	c.Writer.Write([]byte("data: [DONE]\n\n"))
	if ok {
		flusher.Flush()
	}
}
