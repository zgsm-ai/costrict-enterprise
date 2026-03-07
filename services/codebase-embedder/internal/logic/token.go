package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// TokenLogic token生成逻辑
type TokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewTokenLogic 创建TokenLogic实例
func NewTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TokenLogic {
	return &TokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GenerateToken 生成JWT令牌
func (l *TokenLogic) GenerateToken(req *types.TokenRequest) (*types.TokenResponseData, error) {
	if err := l.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 1. 读取限流配置文件
	tokenLimit := l.svcCtx.Config.TokenLimit
	if !tokenLimit.Enabled {
		// 限流未启用，直接生成token
		return l.generateToken(req)
	}

	// 2. 查询任务池正运行任务
	runningTasks, err := l.getRunningTasksCount()
	if err != nil {
		return nil, fmt.Errorf("查询运行中任务失败: %w", err)
	}

	// 3. 判断是否到达限流配置
	if runningTasks >= tokenLimit.MaxRunningTasks {
		// 4. 生成失败
		return nil, types.ErrRateLimitReached
	}

	// 5. 根据ClientId生成Token
	return l.generateToken(req)
}

// generateToken 生成token
func (l *TokenLogic) generateToken(req *types.TokenRequest) (*types.TokenResponseData, error) {
	// 使用clientId和codebasePath生成token
	// 这里使用简单的哈希组合，实际生产环境应使用更安全的JWT实现
	token := fmt.Sprintf("%s_%s_%s", req.ClientId, req.CodebasePath, l.generateRandomString(16))

	return &types.TokenResponseData{
		Token:     token,
		ExpiresIn: 3600, // 1小时 = 3600秒
		TokenType: "Bearer",
	}, nil
}

// getRunningTasksCount 获取运行中任务数量
func (l *TokenLogic) getRunningTasksCount() (int, error) {
	// 获取任务池中正在运行的任务数量
	if l.svcCtx.TaskPool == nil {
		return 0, fmt.Errorf("任务池未初始化")
	}

	return l.svcCtx.TaskPool.Running(), nil
}

// validateRequest 验证请求参数
func (l *TokenLogic) validateRequest(req *types.TokenRequest) error {
	if req.ClientId == "" {
		return errors.New("clientId is required")
	}
	if req.CodebasePath == "" {
		return errors.New("codebasePath is required")
	}
	if req.CodebaseName == "" {
		return errors.New("codebaseName is required")
	}
	return nil
}

// generateRandomString 生成随机字符串
func (l *TokenLogic) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	// 使用更稳定的随机源
	seed := time.Now().UnixNano()
	for i := range result {
		seed = (seed*1103515245 + 12345) & 0x7fffffff
		result[i] = charset[seed%int64(len(charset))]
	}
	return string(result)
}
