package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// RunningTasksLogic 运行中任务查询逻辑
type RunningTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRunningTasksLogic 创建运行中任务查询逻辑
func NewRunningTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunningTasksLogic {
	return &RunningTasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetRunningTasks 获取运行中任务
func (l *RunningTasksLogic) GetRunningTasks() (*types.RunningTasksResponse, error) {
	// 检查Redis连接
	if err := l.checkRedisConnection(); err != nil {
		return nil, fmt.Errorf("Redis服务不可用，请稍后再试")
	}
	
	// 扫描任务状态
	tasks, err := l.scanRunningTasks()
	if err != nil {
		return nil, fmt.Errorf("查询任务状态时发生内部错误: %w", err)
	}
	
	return &types.RunningTasksResponse{
		Code:    0,
		Message: "ok",
		Success: true,
		Data: &types.RunningTasksData{
			TotalTasks: len(tasks),
			Tasks:      tasks,
		},
	}, nil
}

// checkRedisConnection 检查Redis连接
func (l *RunningTasksLogic) checkRedisConnection() error {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()
	
	return l.svcCtx.StatusManager.CheckConnection(ctx)
}

// scanRunningTasks 扫描运行中的任务
func (l *RunningTasksLogic) scanRunningTasks() ([]types.RunningTaskInfo, error) {
	return l.svcCtx.StatusManager.ScanRunningTasks(l.ctx)
}