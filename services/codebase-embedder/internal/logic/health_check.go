package logic

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

type HealthCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthCheckLogic {
	return &HealthCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CheckHealth 执行探活检查
func (l *HealthCheckLogic) CheckHealth(authorization string, req *types.IndexSummaryRequest) error {
	// 如果探活检查未启用，直接返回成功
	if !l.svcCtx.Config.HealthCheck.Enabled {
		l.Info("health check is disabled, skipping")
		return nil
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: l.svcCtx.Config.HealthCheck.Timeout,
	}

	// 构建请求URL
	healthCheckURL := l.svcCtx.Config.HealthCheck.URL

	// 将req参数添加到查询参数中
	if req != nil {
		// 解析基础URL
		parsedURL, err := url.Parse(healthCheckURL)
		if err != nil {
			l.Errorf("failed to parse health check URL: %v", err)
			return fmt.Errorf("failed to parse health check URL: %w", err)
		}

		// 获取现有查询参数或创建新的
		q := parsedURL.Query()

		// 添加req中的参数
		if req.ClientId != "" {
			q.Set("clientId", req.ClientId)
		}
		if req.CodebasePath != "" {
			q.Set("codebasePath", req.CodebasePath)
		}

		// 将查询参数设置回URL
		parsedURL.RawQuery = q.Encode()
		healthCheckURL = parsedURL.String()
	}

	// 创建请求
	httpReq, err := http.NewRequestWithContext(l.ctx, "GET", healthCheckURL, nil)
	if err != nil {
		l.Errorf("failed to create health check request: %v", err)
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Authorization", authorization)
	httpReq.Header.Set("X-Costrict-Version", "v1.0.6")

	// 重试逻辑
	var lastErr error

	// 发送请求
	resp, err := client.Do(httpReq)
	if err != nil {
		lastErr = fmt.Errorf("health check request failed: %w", err)
		return lastErr
	}

	// 确保响应体被关闭
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode == http.StatusOK {
		l.Info("health check passed")
		return nil
	}

	lastErr = fmt.Errorf("health check returned status code: %d", resp.StatusCode)

	return fmt.Errorf("health check failed : %w", lastErr)
}
