package codebase_context

import (
	"bytes"
	"code-completion/pkg/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

// HTTPClient 接口定义
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// APIClient API客户端结构体
type APIClient struct {
	client HTTPClient
}

// NewAPIClient 创建新的API客户端
func NewAPIClient() *APIClient {
	return &APIClient{
		client: &http.Client{
			Timeout: config.Context.RequestTimeout,
		},
	}
}

// RequestParam 请求参数
type RequestParam struct {
	ClientID       string  `json:"clientId"`
	CodebasePath   string  `json:"codebasePath"`
	FilePath       string  `json:"filePath,omitempty"`
	CodeSnippet    string  `json:"codeSnippet,omitempty"`
	StartLine      int     `json:"startLine,omitempty"`
	EndLine        int     `json:"endLine,omitempty"`
	Query          string  `json:"query,omitempty"`
	TopK           int     `json:"topK,omitempty"`
	ScoreThreshold float64 `json:"scoreThreshold,omitempty"`
	MaxLayer       int     `json:"maxLayer,omitempty"`
	IncludeContent bool    `json:"includeContent,omitempty"`
}

// ResponseData 响应数据结构
type ResponseData struct {
	Data struct {
		List []map[string]interface{} `json:"list"`
	} `json:"data"`
}

func kvs2UrlValues(kvs map[string]interface{}) url.Values {
	values := url.Values{}
	for key, value := range kvs {
		switch v := value.(type) {
		case string:
			values.Add(key, v)
		case int, int8, int16, int32, int64:
			values.Add(key, fmt.Sprintf("%d", v))
		case float32, float64:
			values.Add(key, fmt.Sprintf("%f", v))
		case bool:
			values.Add(key, fmt.Sprintf("%t", v))
		case []interface{}:
			// 处理数组类型
			for _, item := range v {
				values.Add(key, fmt.Sprintf("%v", item))
			}
		default:
			if value != nil {
				values.Add(key, fmt.Sprintf("%v", value))
			}
		}
	}
	return values
}

func headers2zapAny(headers http.Header) map[string]interface{} {
	headerMap := make(map[string]interface{})
	for key, values := range headers {
		headerMap[key] = values
	}
	return headerMap
}

// doRequest 发送HTTP请求
func (c *APIClient) DoRequest(ctx context.Context, requestURL string, params RequestParam, headers http.Header, method string) (*ResponseData, error) {
	var req *http.Request
	body, err := json.Marshal(params)
	if err != nil {
		zap.L().Warn("Failed to marshal request params", zap.Error(err), zap.String("url", requestURL))
		return nil, err
	}
	if method == "POST" {
		req, err = http.NewRequestWithContext(ctx, method, requestURL, bytes.NewBuffer(body))
	} else {
		var paramsMap map[string]interface{}
		if err := json.Unmarshal(body, &paramsMap); err != nil {
			zap.L().Warn("Failed to unmarshal params for query", zap.Error(err), zap.String("url", requestURL))
			return nil, err
		}
		query := kvs2UrlValues(paramsMap)
		rawUrl := requestURL
		if len(query) > 0 {
			rawUrl = requestURL + "?" + query.Encode()
		}
		req, err = http.NewRequestWithContext(ctx, method, rawUrl, nil)
	}
	if err != nil {
		zap.L().Warn("Failed to create request", zap.Error(err), zap.String("url", requestURL))
		return nil, err
	}
	// 设置请求头,只包含这几个
	req.Header.Set("X-Request-Id", headers.Get("X-Request-Id"))
	req.Header.Set("Authorization", headers.Get("Authorization"))
	req.Header.Set("X-Costrict-Version", headers.Get("X-Costrict-Version"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		zap.L().Warn("Request failed", zap.Error(err),
			zap.String("url", requestURL),
			zap.String("body", string(body)))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		zap.L().Warn("Request returned non-200 status",
			zap.Int("status", resp.StatusCode),
			zap.String("url", requestURL),
			zap.Any("headers", headers2zapAny(req.Header)),
			zap.String("params", string(body)),
			zap.String("resp", string(data)))
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}
	var result ResponseData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		zap.L().Warn("Failed to decode response", zap.Error(err), zap.String("url", requestURL))
		return nil, err
	}
	return &result, nil
}
