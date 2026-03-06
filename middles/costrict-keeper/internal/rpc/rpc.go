package rpc

import (
	"bytes"
	"costrict-keeper/internal/env"
	"costrict-keeper/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HTTPClient 定义HTTP客户端接口
type HTTPClient interface {
	Get(path string, params map[string]interface{}) (*HTTPResponse, error)
	Post(path string, data interface{}) (*HTTPResponse, error)
	Put(path string, data interface{}) (*HTTPResponse, error)
	Patch(path string, data interface{}) (*HTTPResponse, error)
	Delete(path string, params map[string]interface{}) (*HTTPResponse, error)
	Close() error
}

// HTTPConfig 定义HTTP客户端配置
type HTTPConfig struct {
	Address string        //costrict服务侦听地址
	Network string        //unix,tcp...
	Timeout time.Duration // 默认超时时间
	BaseURL string        // 基础URL
}

// DefaultHTTPConfig 返回默认HTTP客户端配置
func DefaultHTTPConfig() *HTTPConfig {
	c := &HTTPConfig{
		Address: getSocketPath("costrict.sock", ""),
		Network: "unix",
		Timeout: 5 * time.Second,
		BaseURL: "http://localhost",
	}
	// 检查socket文件是否存在
	if _, err := os.Stat(c.Address); os.IsNotExist(err) {
		c.Address = getTcpAddress()
		c.Network = "tcp"
	}
	if c.Address == "" {
		c.Address = "127.0.0.1:8999"
		c.Network = "tcp"
	}
	return c
}

// HTTPResponse 定义HTTP响应结构
type HTTPResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
	Error      string              `json:"error"`
}

// buildURL 构建完整的URL
func buildURL(baseURL, path string, params map[string]interface{}) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	// 添加路径
	if u.Path == "" {
		u.Path = path
	} else {
		// 确保路径以/结尾，然后拼接
		if !strings.HasSuffix(u.Path, "/") {
			u.Path += "/"
		}
		u.Path += path
	}

	// 添加查询参数
	if params != nil {
		q := u.Query()
		for key, value := range params {
			switch v := value.(type) {
			case string:
				q.Set(key, v)
			case int, int8, int16, int32, int64:
				q.Set(key, fmt.Sprintf("%d", v))
			case uint, uint8, uint16, uint32, uint64:
				q.Set(key, fmt.Sprintf("%d", v))
			case float32, float64:
				q.Set(key, fmt.Sprintf("%f", v))
			case bool:
				q.Set(key, fmt.Sprintf("%t", v))
			default:
				q.Set(key, fmt.Sprintf("%v", v))
			}
		}
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}

// serializeData 序列化请求数据
func serializeData(data interface{}) (io.Reader, error) {
	if data == nil {
		return nil, nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}

	return bytes.NewReader(jsonData), nil
}

// deserializeResponse 反序列化响应数据
func deserializeResponse(resp *http.Response) (*HTTPResponse, error) {
	httpResp := &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	httpResp.Body = body
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return httpResp, nil
	}
	if len(body) == 0 {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			httpResp.Error = resp.Status
		}
	} else {
		var errBody models.ErrorResponse
		if err := json.Unmarshal(body, &errBody); err != nil {
			httpResp.Error = err.Error()
		} else {
			httpResp.Error = errBody.Error
		}
	}
	if httpResp.Error == "" {
		httpResp.Error = "Unknown error"
	}
	return httpResp, nil
}

/**
 * costrict服务侦听的unix socket地址
 */
func getSocketPath(socketName string, socketDir string) string {
	if socketDir == "" {
		socketDir = filepath.Join(env.CostrictDir, "run")
	}
	return filepath.Join(socketDir, socketName)
}

/**
 * costrict服务侦听的tcp地址
 */
func getTcpAddress() string {
	knownFile := filepath.Join(env.CostrictDir, "share", ".well-known.json")
	data, err := os.ReadFile(knownFile)
	if err != nil {
		return ""
	}
	var known models.SystemKnowledge
	if err = json.Unmarshal(data, &known); err != nil {
		return ""
	}
	for _, s := range known.Services {
		if s.Name == "costrict" {
			return fmt.Sprintf("127.0.0.1:%d", s.Port)
		}
	}
	return ""
}
