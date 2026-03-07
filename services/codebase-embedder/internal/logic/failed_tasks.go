package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// FailedTasksLogic 失败任务查询逻辑
type FailedTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFailedTasksLogic 创建失败任务查询逻辑
func NewFailedTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FailedTasksLogic {
	return &FailedTasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetFailedTasks 获取失败任务
func (l *FailedTasksLogic) GetFailedTasks() (*types.FailedTasksResponse, error) {
	// 检查Redis连接
	if err := l.checkRedisConnection(); err != nil {
		return nil, fmt.Errorf("Redis服务不可用，请稍后再试")
	}

	// 扫描失败任务
	tasks, err := l.scanFailedTasks()
	if err != nil {
		return nil, fmt.Errorf("查询失败任务时发生内部错误: %w", err)
	}

	return &types.FailedTasksResponse{
		Code:    0,
		Message: "ok",
		Success: true,
		Data: &types.FailedTasksData{
			TotalTasks: len(tasks),
			Tasks:      tasks,
		},
	}, nil
}

// checkRedisConnection 检查Redis连接
func (l *FailedTasksLogic) checkRedisConnection() error {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	return l.svcCtx.StatusManager.CheckConnection(ctx)
}

// scanFailedTasks 扫描失败的任务
func (l *FailedTasksLogic) scanFailedTasks() ([]types.CompletedTaskInfo, error) {
	return l.svcCtx.StatusManager.ScanFailedTasks(l.ctx)
}
