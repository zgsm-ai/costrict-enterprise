package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zgsm-ai/codebase-indexer/internal/config"
)

// PortResponse 接口响应结构
type PortResponse struct {
	Port int `json:"mappingPort"`
}

// PortManager 端口管理器
type PortManager struct {
	baseURL    string
	forwardURL string
	httpClient *http.Client
	cache      map[string]PortResponse
	cacheExp   time.Duration
	lastUpdate map[string]time.Time
	mu         sync.RWMutex
}

// NewPortManager 创建端口管理器
func NewPortManager(baseURL string) *PortManager {
	return &PortManager{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		forwardURL: "http://10.233.23.31", // 默认转发地址
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		cache:      make(map[string]PortResponse),
		cacheExp:   5 * time.Minute, // 缓存5分钟
		lastUpdate: make(map[string]time.Time),
	}
}

// NewPortManagerWithConfig 从配置创建端口管理器
func NewPortManagerWithConfig(config config.PortManagerConfig) *PortManager {
	// 设置默认值
	baseURL := config.URL
	forwardURL := config.ForwardURL
	if forwardURL == "" {
		forwardURL = "http://10.233.23.31" // 默认转发地址
	}
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	cacheExp := config.CacheExp
	if cacheExp == 0 {
		cacheExp = 5 * time.Minute
	}
	maxIdleConns := config.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 10
	}
	maxIdleConnsPerHost := config.MaxIdleConnsPerHost
	if maxIdleConnsPerHost == 0 {
		maxIdleConnsPerHost = 5
	}
	idleConnTimeout := config.IdleConnTimeout
	if idleConnTimeout == 0 {
		idleConnTimeout = 30 * time.Second
	}

	return &PortManager{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		forwardURL: strings.TrimSuffix(forwardURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        maxIdleConns,
				MaxIdleConnsPerHost: maxIdleConnsPerHost,
				IdleConnTimeout:     idleConnTimeout,
			},
		},
		cache:      make(map[string]PortResponse),
		cacheExp:   cacheExp,
		lastUpdate: make(map[string]time.Time),
	}
}

// GetPort 获取端口信息
func (pm *PortManager) GetPort(ctx context.Context, clientID, appName string, headers http.Header) (*PortResponse, error) {
	cacheKey := fmt.Sprintf("%s:%s", clientID, appName)

	// 检查缓存
	pm.mu.RLock()
	if cached, exists := pm.cache[cacheKey]; exists {
		if time.Since(pm.lastUpdate[cacheKey]) < pm.cacheExp {
			pm.mu.RUnlock()
			logx.Infof("Using cached port for client %s, app %s: %d", clientID, appName, cached.Port)
			return &cached, nil
		}
	}
	pm.mu.RUnlock()

	// 构建请求URL
	requestURL := fmt.Sprintf("%s/tunnel-manager/api/v1/ports?clientId=%s&appName=%s",
		pm.baseURL, clientID, appName)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 复制原始请求头到端口管理器请求中
	if headers != nil {
		for key, values := range headers {
			// 跳过可能冲突的头部
			if strings.ToLower(key) == "host" || strings.ToLower(key) == "content-length" {
				continue
			}
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	// 发送请求
	resp, err := pm.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch port: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var portResp PortResponse
	if err := json.NewDecoder(resp.Body).Decode(&portResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 打印详细的请求和响应信息
	headersStr := ""
	for key, values := range req.Header {
		headersStr += fmt.Sprintf("%s: %s; ", key, strings.Join(values, ","))
	}
	logx.Infof("Port request completed - URL: %s, Headers: [%s], StatusCode: %d, Response: %+v", 
		requestURL, headersStr, resp.StatusCode, portResp)

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 更新缓存
	pm.mu.Lock()
	pm.cache[cacheKey] = portResp
	pm.lastUpdate[cacheKey] = time.Now()
	pm.mu.Unlock()

	logx.Infof("Successfully fetched port for client %s, app %s: %d", clientID, appName, portResp.Port)
	return &portResp, nil
}

// GetPortFromHeaders 从请求获取端口信息
// 对于 GET 请求，从 params 中获取 clientId
// 对于其他请求，从 body 中获取 clientId
func (pm *PortManager) GetPortFromHeaders(ctx context.Context, method string, headers http.Header, params map[string][]string, body []byte) (*PortResponse, error) {
	var clientID string

	// 根据请求方法从不同位置获取 clientId
	if method == "GET" {
		// 从 GET 请求的 params 中获取 clientId
		if clientIds, exists := params["clientId"]; exists && len(clientIds) > 0 {
			clientID = clientIds[0]
		}
	} else {
		// 从其他请求的 body 中获取 clientId
		if len(body) > 0 {
			// 尝试解析 JSON body
			var requestBody map[string]interface{}
			if err := json.Unmarshal(body, &requestBody); err == nil {
				if id, exists := requestBody["clientId"]; exists {
					if idStr, ok := id.(string); ok {
						clientID = idStr
					}
				}
			}
		}
	}

	if clientID == "" {
		// 如果从 params 或 body 中获取不到，尝试从 headers 中获取（向后兼容）
		clientID = headers.Get("clientId")
		if clientID == "" {
			return nil, fmt.Errorf("clientId is required in params (for GET) or body (for other methods) or headers")
		}
	}

	appName := "codebase-indexer"

	return pm.GetPort(ctx, clientID, appName, headers)
}

// BuildTargetURL 构建目标URL
func (pm *PortManager) BuildTargetURL(portResp *PortResponse) string {
	// 构建完整URL
	return fmt.Sprintf("%s:%d", pm.forwardURL, portResp.Port)
}
