package handler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/client"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"go.uber.org/zap"
)

// ForwardHandler handles request forwarding
func ForwardHandler(svcCtx *bootstrap.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if forwarding is enabled
		if !svcCtx.Config.Forward.Enabled {
			sendErrorResponse(c, http.StatusForbidden, fmt.Errorf("forwarding is disabled"))
			return
		}

		// Record the start time
		startTime := time.Now()

		// Get the target URL from query parameters or configuration
		targetURL := c.Query("target")
		if targetURL == "" {
			targetURL = svcCtx.Config.Forward.DefaultTarget
		}

		// If not provided in query, use default from config if available
		if targetURL != "" {
			// 将 defaultTarget 作为 baseURL，拼接当前请求的 path
			baseURL := strings.TrimSuffix(targetURL, "/")
			// 提取 /chat-rag/api/forward/ 之后的 path
			path := strings.TrimPrefix(c.Request.URL.Path, "/chat-rag/api/forward")
			path = strings.TrimPrefix(path, "/")
			
			// 根据path是否为空决定如何拼接URL
			if path != "" {
				targetURL = baseURL + "/" + path
			} else {
				targetURL = baseURL
			}
		} else {
			sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("target URL is required"))
			return
		}

		// Log the incoming request
		logIncomingRequest(c, targetURL)

		// Read the request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			// Restore the body so it can be read again
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Request body logging is disabled

		// Create HTTP client with 10 minutes timeout
		httpClient := client.NewHTTPClient(targetURL, client.HTTPClientConfig{
			Timeout: 10 * time.Minute,
		})

		// Prepare the request
		req := client.Request{
			Method:  c.Request.Method,
			Path:    "", // The target URL already includes the full path
			Headers: make(map[string]string),
			Body:    nil,
		}

		// Copy headers from the original request
		for key, values := range c.Request.Header {
			if len(values) > 0 {
				req.Headers[key] = values[0]
			}
		}

		// Set the request body if present
		if len(bodyBytes) > 0 {
			// Try to unmarshal as JSON first, if that fails, treat as raw bytes
			var jsonBody interface{}
			if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
				req.Body = jsonBody
			} else {
				req.Body = string(bodyBytes)
			}
		}

		// Forward the request
		resp, err := httpClient.DoRequest(c.Request.Context(), req)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(c.Request.Context().Err(), context.Canceled) {
				logger.Warn("Forward request canceled by client",
					zap.String("targetURL", targetURL),
					zap.Error(err),
				)
				return
			}

			logger.Error("Failed to forward request",
				zap.String("targetURL", targetURL),
				zap.Error(err),
			)

			// Save the forward log with error
			if saveErr := saveForwardLog(c, targetURL, nil, nil, time.Since(startTime), svcCtx.Config.Log.LogFilePath, err); saveErr != nil {
				logger.Error("Failed to save forward log",
					zap.String("targetURL", targetURL),
					zap.Error(saveErr),
				)
			}

			sendErrorResponse(c, http.StatusBadGateway, fmt.Errorf("failed to forward request: %w", err))
			return
		}
		defer resp.Body.Close()

		// Read the response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("Failed to read response body",
				zap.String("targetURL", targetURL),
				zap.Error(err),
			)

			// Save the forward log with error
			if saveErr := saveForwardLog(c, targetURL, resp, nil, time.Since(startTime), svcCtx.Config.Log.LogFilePath, err); saveErr != nil {
				logger.Error("Failed to save forward log",
					zap.String("targetURL", targetURL),
					zap.Error(saveErr),
				)
			}

			sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to read response: %w", err))
			return
		}

		// Log the response
		logger.Info("Forward response",
			zap.String("targetURL", targetURL),
			zap.Int("statusCode", resp.StatusCode),
			zap.Duration("duration", time.Since(startTime)),
		)

		// Save the forward log to file
		if err := saveForwardLog(c, targetURL, resp, respBody, time.Since(startTime), svcCtx.Config.Log.LogFilePath, nil); err != nil {
			logger.Error("Failed to save forward log",
				zap.String("targetURL", targetURL),
				zap.Error(err),
			)
		}

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Set the status code
		c.Status(resp.StatusCode)

		// Write the response body
		if _, err := c.Writer.Write(respBody); err != nil {
			logger.Error("Failed to write response",
				zap.String("targetURL", targetURL),
				zap.Error(err),
			)
		}
	}
}

// logIncomingRequest logs the details of the incoming request
func logIncomingRequest(c *gin.Context, targetURL string) {
	logger.Info("Forwarding request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("query", c.Request.URL.RawQuery),
		zap.String("targetURL", targetURL),
	)
}

// saveForwardLog saves the forward request and response to a log file
func saveForwardLog(c *gin.Context, targetURL string, resp *http.Response, respBody []byte, duration time.Duration, logFilePath string, requestErr error) error {
	// Read the request body again
	var bodyBytes []byte
	if c.Request.Body != nil {
		// We need to restore the body after reading
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Parse request body as JSON if possible - always expand compressed JSON
	var reqBody interface{}
	if len(bodyBytes) > 0 {
		// Check if the request body is gzip compressed
		var decodedReqBody string
		if isGzipData(bodyBytes) {
			// Decompress gzip data
			decompressed, err := decompressGzip(bodyBytes)
			if err == nil {
				decodedReqBody = decompressed
			} else {
				// If decompression fails, fall back to original
				decodedReqBody = decodeUnicodeEscapes(string(bodyBytes))
			}
		} else {
			// Not compressed, just decode Unicode escapes
			decodedReqBody = decodeUnicodeEscapes(string(bodyBytes))
		}

		// Debug: Log the decoded request body
		logger.Debug("Decoded request body",
			zap.String("decodedBody", decodedReqBody),
		)

		// Try to parse as JSON first
		if err := json.Unmarshal([]byte(decodedReqBody), &reqBody); err != nil {
			// If JSON parsing failed, log the error and store as string
			logger.Debug("Failed to parse request body as JSON",
				zap.String("error", err.Error()),
				zap.String("body", decodedReqBody),
			)
			reqBody = decodedReqBody
		} else {
			// If JSON parsing succeeded, recursively decode Unicode escapes in the JSON structure
			// This ensures the JSON is fully expanded and properly formatted
			reqBody = decodeUnicodeEscapesInJSON(reqBody)

			// Debug: Log the parsed JSON structure
			logger.Debug("Successfully parsed request body as JSON",
				zap.Any("parsedBody", reqBody),
			)
		}
	}

	// Parse response body - simplify logic for JSON vs SSE
	var resBody interface{}
	var beautyBody interface{}
	var bodyContent string

	if resp != nil && len(respBody) > 0 {
		respBodyStr := string(respBody)

		// Check if the response is gzip compressed
		var decodedRespBody string
		if isGzipData(respBody) {
			// Decompress gzip data
			decompressed, err := decompressGzip(respBody)
			if err == nil {
				decodedRespBody = decompressed
			} else {
				// If decompression fails, fall back to original
				decodedRespBody = decodeUnicodeEscapes(respBodyStr)
			}
		} else {
			// Not compressed, just decode Unicode escapes
			decodedRespBody = decodeUnicodeEscapes(respBodyStr)
		}

		// Check if it's SSE (Server-Sent Events) format
		if isSSEFormat(decodedRespBody) {
			// For SSE format, keep existing logic
			beautyBody = formatSSEContent(decodedRespBody)

			// Extract content from SSE data blocks for body_content only if it's event/data format
			if hasEventAndDataFormat(decodedRespBody) {
				bodyContent = extractContentFromSSE(decodedRespBody)
			} else {
				bodyContent = ""
			}

			// For SSE, resBody remains as original decoded string
			resBody = decodedRespBody
		} else {
			// For JSON format, directly expand and don't use beauty_body/body_content
			var respBodyObj interface{}
			if err := json.Unmarshal([]byte(decodedRespBody), &respBodyObj); err == nil {
				// If it's valid JSON, expand it properly
				resBody = decodeUnicodeEscapesInJSON(respBodyObj)
			} else {
				// If not JSON, store as string
				resBody = decodedRespBody
			}

			// For non-SSE responses, don't set beauty_body and body_content
			beautyBody = nil
			bodyContent = ""
		}
	}

	// Create forward log entry
	forwardLog := model.ForwardLog{
		Timestamp: time.Now(),
		TargetURL: targetURL,
		Duration:  duration,
		Request: model.ForwardRequest{
			Method:  c.Request.Method,
			Path:    c.Request.URL.Path,
			Query:   c.Request.URL.RawQuery,
			Headers: make(map[string]string),
			Body:    reqBody,
		},
		Response: model.ForwardResponse{
			StatusCode:  0,
			Headers:     make(map[string]string),
			Body:        resBody,
			BeautyBody:  beautyBody,  // Will be nil for JSON responses
			BodyContent: bodyContent, // Will be empty for JSON responses
		},
	}

	// If resp is not nil, set status code
	if resp != nil {
		forwardLog.Response.StatusCode = resp.StatusCode
	}

	// Add error message if present
	if requestErr != nil {
		forwardLog.Error = requestErr.Error()
	}

	// Copy request headers
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			forwardLog.Request.Headers[key] = values[0]
		}
	}

	// Copy response headers (only if resp is not nil)
	if resp != nil {
		for key, values := range resp.Header {
			if len(values) > 0 {
				forwardLog.Response.Headers[key] = values[0]
			}
		}
	}

	// Convert log to JSON
	logJSON, err := forwardLog.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal forward log: %w", err)
	}

	// Create logs directory if it doesn't exist - format: forward/年-月-日/
	dateStr := time.Now().Format("2006-01-02")
	forwardLogDir := filepath.Join(logFilePath, "forward", dateStr)
	if err := os.MkdirAll(forwardLogDir, 0755); err != nil {
		return fmt.Errorf("failed to create forward logs directory: %w", err)
	}

	// Generate filename with timestamp and path
	timestamp := time.Now().Format("20060102-150405")
	// Extract path from forward URL, remove leading slash and replace slashes with underscores
	pathPart := strings.TrimPrefix(c.Request.URL.Path, "/chat-rag/api/forward/")
	pathPart = strings.ReplaceAll(pathPart, "/", "_")
	// Generate 6-digit random number
	randomNum := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	filename := fmt.Sprintf("%s-%s-%s.json", timestamp, pathPart, randomNum)
	filePath := filepath.Join(forwardLogDir, filename)

	// Write log to file
	if err := os.WriteFile(filePath, []byte(logJSON), 0644); err != nil {
		return fmt.Errorf("failed to write forward log file: %w", err)
	}

	logger.Info("Forward log saved",
		zap.String("filePath", filePath),
		zap.String("targetURL", targetURL),
	)

	return nil
}

// isSSEFormat checks if the response body is in Server-Sent Events (SSE) format
func isSSEFormat(content string) bool {
	return strings.Contains(content, "data: ") && strings.Contains(content, "\n\n")
}

// decodeUnicodeEscapesInJSON recursively decodes Unicode escape sequences in JSON structure
func decodeUnicodeEscapesInJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case string:
		// If it's a string, decode Unicode escapes
		return decodeUnicodeEscapes(v)
	case map[string]interface{}:
		// If it's a map, recursively process each value
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = decodeUnicodeEscapesInJSON(value)
		}
		return result
	case []interface{}:
		// If it's an array, recursively process each element
		result := make([]interface{}, len(v))
		for i, value := range v {
			result[i] = decodeUnicodeEscapesInJSON(value)
		}
		return result
	default:
		// For other types (numbers, booleans, null), return as-is
		return data
	}
}

// decodeUnicodeEscapes decodes Unicode escape sequences in a string
func decodeUnicodeEscapes(s string) string {
	// Debug: Log input string
	logger.Debug("decodeUnicodeEscapes input",
		zap.String("input", s),
	)

	// Direct string replacement for common Unicode escape sequences
	result := s

	// HTML/XML special characters - handle both single and double backslash patterns
	result = strings.ReplaceAll(result, "\\u003c", "<")  // <
	result = strings.ReplaceAll(result, "\u003c", "<")   // < (single backslash)
	result = strings.ReplaceAll(result, "\\u003e", ">")  // >
	result = strings.ReplaceAll(result, "\u003e", ">")   // > (single backslash)
	result = strings.ReplaceAll(result, "\\u0026", "&")  // &
	result = strings.ReplaceAll(result, "\u0026", "&")   // & (single backslash)
	result = strings.ReplaceAll(result, "\\u0022", "\"") // "
	result = strings.ReplaceAll(result, "\u0022", "\"")  // " (single backslash)
	result = strings.ReplaceAll(result, "\\u0027", "'")  // '
	result = strings.ReplaceAll(result, "\u0027", "'")   // ' (single backslash)

	// Common punctuation and symbols - handle both single and double backslash patterns
	result = strings.ReplaceAll(result, "\\u003d", "=")  // =
	result = strings.ReplaceAll(result, "\u003d", "=")   // = (single backslash)
	result = strings.ReplaceAll(result, "\\u002b", "+")  // +
	result = strings.ReplaceAll(result, "\u002b", "+")   // + (single backslash)
	result = strings.ReplaceAll(result, "\\u002d", "-")  // -
	result = strings.ReplaceAll(result, "\u002d", "-")   // - (single backslash)
	result = strings.ReplaceAll(result, "\\u002a", "*")  // *
	result = strings.ReplaceAll(result, "\u002a", "*")   // * (single backslash)
	result = strings.ReplaceAll(result, "\\u002f", "/")  // /
	result = strings.ReplaceAll(result, "\u002f", "/")   // / (single backslash)
	result = strings.ReplaceAll(result, "\\u005c", "\\") // \
	result = strings.ReplaceAll(result, "\u005c", "\\")  // \ (single backslash)
	result = strings.ReplaceAll(result, "\\u007c", "|")  // |
	result = strings.ReplaceAll(result, "\u007c", "|")   // | (single backslash)

	// Brackets and braces - handle both single and double backslash patterns
	result = strings.ReplaceAll(result, "\\u0028", "(") // (
	result = strings.ReplaceAll(result, "\u0028", "(")  // ( (single backslash)
	result = strings.ReplaceAll(result, "\\u0029", ")") // )
	result = strings.ReplaceAll(result, "\u0029", ")")  // ) (single backslash)
	result = strings.ReplaceAll(result, "\\u005b", "[") // [
	result = strings.ReplaceAll(result, "\u005b", "[")  // [ (single backslash)
	result = strings.ReplaceAll(result, "\\u005d", "]") // ]
	result = strings.ReplaceAll(result, "\u005d", "]")  // ] (single backslash)
	result = strings.ReplaceAll(result, "\\u007b", "{") // {
	result = strings.ReplaceAll(result, "\u007b", "{")  // { (single backslash)
	result = strings.ReplaceAll(result, "\\u007d", "}") // }
	result = strings.ReplaceAll(result, "\u007d", "}")  // } (single backslash)

	// Whitespace characters - handle both single and double backslash patterns
	result = strings.ReplaceAll(result, "\\u0020", " ")  // space
	result = strings.ReplaceAll(result, "\u0020", " ")   // space (single backslash)
	result = strings.ReplaceAll(result, "\\u0009", "\t") // tab
	result = strings.ReplaceAll(result, "\u0009", "\t")  // tab (single backslash)
	result = strings.ReplaceAll(result, "\\u000a", "\n") // line feed
	result = strings.ReplaceAll(result, "\u000a", "\n")  // line feed (single backslash)
	result = strings.ReplaceAll(result, "\\u000d", "\r") // carriage return
	result = strings.ReplaceAll(result, "\u000d", "\r")  // carriage return (single backslash)

	// Other common symbols - handle both single and double backslash patterns
	result = strings.ReplaceAll(result, "\\u0021", "!") // !
	result = strings.ReplaceAll(result, "\u0021", "!")  // ! (single backslash)
	result = strings.ReplaceAll(result, "\\u003f", "?") // ?
	result = strings.ReplaceAll(result, "\u003f", "?")  // ? (single backslash)
	result = strings.ReplaceAll(result, "\\u0023", "#") // #
	result = strings.ReplaceAll(result, "\u0023", "#")  // # (single backslash)
	result = strings.ReplaceAll(result, "\\u0024", "$") // $
	result = strings.ReplaceAll(result, "\u0024", "$")  // $ (single backslash)
	result = strings.ReplaceAll(result, "\\u0025", "%") // %
	result = strings.ReplaceAll(result, "\u0025", "%")  // % (single backslash)
	result = strings.ReplaceAll(result, "\\u0040", "@") // @
	result = strings.ReplaceAll(result, "\u0040", "@")  // @ (single backslash)
	result = strings.ReplaceAll(result, "\\u005e", "^") // ^
	result = strings.ReplaceAll(result, "\u005e", "^")  // ^ (single backslash)
	result = strings.ReplaceAll(result, "\\u0060", "`") // `
	result = strings.ReplaceAll(result, "\u0060", "`")  // ` (single backslash)
	result = strings.ReplaceAll(result, "\\u007e", "~") // ~
	result = strings.ReplaceAll(result, "\u007e", "~")  // ~ (single backslash)

	logger.Debug("decodeUnicodeEscapes output",
		zap.String("output", result),
	)

	return result
}

// decompressGzip decompresses gzip compressed data
func decompressGzip(data []byte) (string, error) {
	// Create a gzip reader
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer r.Close()

	// Read the decompressed data
	decompressed, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	return string(decompressed), nil
}

// isGzipData checks if the data is gzip compressed
func isGzipData(data []byte) bool {
	// Gzip magic number: 0x1f 0x8b
	if len(data) < 2 {
		return false
	}
	return data[0] == 0x1f && data[1] == 0x8b
}

// hasEventAndDataFormat checks if the SSE content contains event and data format
func hasEventAndDataFormat(content string) bool {
	// Split by SSE data blocks
	blocks := strings.Split(content, "\n\n")

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		// Split block into lines
		lines := strings.Split(block, "\n")
		hasEvent := false
		hasData := false

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "event: ") {
				hasEvent = true
			} else if strings.HasPrefix(line, "data: ") {
				hasData = true
			}
		}

		// If this block has both event and data, return true (Claude format)
		if hasEvent && hasData {
			return true
		}

		// If this block has only data, check if it's OpenAI format
		if hasData && !hasEvent {
			// Try to find a data line and check if it contains OpenAI format
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "data: ") {
					dataStr := strings.TrimPrefix(line, "data: ")

					// Skip [DONE] marker
					if dataStr == "[DONE]" {
						continue
					}

					// Try to parse JSON
					var jsonData map[string]interface{}
					if err := json.Unmarshal([]byte(dataStr), &jsonData); err == nil {
						// Check if it's OpenAI format (has choices array)
						if _, ok := jsonData["choices"]; ok {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// extractContentFromSSE extracts and concatenates content from SSE data blocks
func extractContentFromSSE(content string) string {
	// Split by SSE data blocks
	blocks := strings.Split(content, "\n\n")
	var contentBuilder strings.Builder

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		// Split block into lines to handle event and data separately
		lines := strings.Split(block, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// 只处理 data: 开头的行，忽略 event: 开头的行
			if strings.HasPrefix(line, "data: ") {
				// Data line, try to parse JSON and extract text content
				dataStr := strings.TrimPrefix(line, "data: ")

				// First decode any Unicode escape sequences in the data string
				decodedDataStr := decodeUnicodeEscapes(dataStr)

				// Try to parse JSON
				var jsonData map[string]interface{}
				if err := json.Unmarshal([]byte(decodedDataStr), &jsonData); err == nil {
					// Successfully parsed as JSON, try to extract text content
					if extractTextFromJSON(jsonData, &contentBuilder) {
						// Content extracted successfully
					}
				} else {
					// Not valid JSON, check if it's [DONE] marker
					if decodedDataStr != "[DONE]" {
						// Add as plain text if not [DONE]
						contentBuilder.WriteString(decodedDataStr)
					}
				}
			}
			// 忽略 event: 开头的行，不拼接到 contentBuilder
		}
	}

	return contentBuilder.String()
}

// extractTextFromJSON recursively extracts text content from JSON structure
func extractTextFromJSON(data interface{}, builder *strings.Builder) bool {
	switch v := data.(type) {
	case map[string]interface{}:
		// 首先尝试适配 OpenAI 格式：data.choices[0].delta.content
		if choices, ok := v["choices"].([]interface{}); ok && len(choices) > 0 {
			if firstChoice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := firstChoice["delta"].(map[string]interface{}); ok {
					// 提取 content 字段
					if content, ok := delta["content"].(string); ok && content != "" {
						// Decode Unicode escapes before adding to builder
						decodedContent := decodeUnicodeEscapes(content)
						builder.WriteString(decodedContent)
					}

					// 提取 tool_calls 字段
					if toolCalls, ok := delta["tool_calls"].([]interface{}); ok {
						for _, toolCall := range toolCalls {
							if toolCallMap, ok := toolCall.(map[string]interface{}); ok {
								if function, ok := toolCallMap["function"].(map[string]interface{}); ok {
									if name, ok := function["name"].(string); ok {
										builder.WriteString(fmt.Sprintf("[Tool Call: %s", name))
										if args, ok := function["arguments"].(string); ok {
											builder.WriteString(fmt.Sprintf(" with arguments: %s", args))
										}
										builder.WriteString("]")
									}
								}
							}
						}
					}

					// 如果有 content 或 tool_calls，则返回 true
					if _, hasContent := delta["content"]; hasContent {
						return true
					}
					if _, hasToolCalls := delta["tool_calls"]; hasToolCalls {
						return true
					}
				}
			}
		}

		// 然后尝试适配 Claude 格式：data.delta.text
		if delta, ok := v["delta"].(map[string]interface{}); ok {
			if textDelta, ok := delta["text"].(string); ok {
				// Decode Unicode escapes before adding to builder
				decodedTextDelta := decodeUnicodeEscapes(textDelta)
				builder.WriteString(decodedTextDelta)
				return true
			}
		}

		// 不再递归处理其他字段，确保只提取 delta.text 或 choices[0].delta.content/tool_calls
		return true

	case []interface{}:
		// Process array elements
		for _, item := range v {
			extractTextFromJSON(item, builder)
		}
		return true

	case string:
		// Direct string value - decode Unicode escapes before adding to builder
		decodedString := decodeUnicodeEscapes(v)
		builder.WriteString(decodedString)
		return true

	default:
		// Other types, ignore
		return false
	}
}

// formatSSEContent formats SSE content into a structured JSON format
func formatSSEContent(content string) interface{} {
	// First, decode any Unicode escape sequences in the entire content
	decodedContent := decodeUnicodeEscapes(content)

	// Split by SSE data blocks
	blocks := strings.Split(decodedContent, "\n\n")
	var events []model.SSEEvent

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		// Create a new event for this block
		event := model.SSEEvent{}

		// Split block into lines to handle event and data separately
		lines := strings.Split(block, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "event: ") {
				// Event line
				event.Event = strings.TrimPrefix(line, "event: ")
			} else if strings.HasPrefix(line, "data: ") {
				// Data line, try to parse JSON
				dataStr := strings.TrimPrefix(line, "data: ")

				// Try to parse JSON
				var jsonData interface{}
				if err := json.Unmarshal([]byte(dataStr), &jsonData); err == nil {
					// Successfully parsed as JSON
					event.Data = jsonData
				} else {
					// Not valid JSON, keep as string but decode any Unicode escapes
					event.Data = decodeUnicodeEscapes(dataStr)
				}
			}
		}

		// Add the event to the list if it has content
		if event.Event != "" || event.Data != nil {
			events = append(events, event)
		}
	}

	return events
}
