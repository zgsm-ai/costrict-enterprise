package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// CompletedTasksLogic 已完成任务查询逻辑
type CompletedTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCompletedTasksLogic 创建已完成任务查询逻辑
func NewCompletedTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompletedTasksLogic {
	return &CompletedTasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetCompletedTasks 获取已完成任务
func (l *CompletedTasksLogic) GetCompletedTasks() (*types.CompletedTasksResponse, error) {
	// 检查Redis连接
	if err := l.checkRedisConnection(); err != nil {
		return nil, fmt.Errorf("Redis服务不可用，请稍后再试")
	}

	// 扫描已完成任务
	tasks, err := l.scanCompletedTasks()
	if err != nil {
		return nil, fmt.Errorf("查询已完成任务时发生内部错误: %w", err)
	}

	return &types.CompletedTasksResponse{
		Code:    0,
		Message: "ok",
		Success: true,
		Data: &types.CompletedTasksData{
			TotalTasks: len(tasks),
			Tasks:      tasks,
		},
	}, nil
}

// checkRedisConnection 检查Redis连接
func (l *CompletedTasksLogic) checkRedisConnection() error {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	return l.svcCtx.StatusManager.CheckConnection(ctx)
}

// scanCompletedTasks 扫描已完成的任务
func (l *CompletedTasksLogic) scanCompletedTasks() ([]types.CompletedTaskInfo, error) {
	return l.svcCtx.StatusManager.ScanCompletedTasks(l.ctx)
}
