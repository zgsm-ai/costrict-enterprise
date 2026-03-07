package logic

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/dao/model"
	"github.com/zgsm-ai/codebase-indexer/internal/tracer"

	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VectorsSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVectorsSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VectorsSummaryLogic {
	return &VectorsSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VectorsSummaryLogic) VectorsSummary(req *types.GetAllVectorsSummaryRequest) (*types.GetAllVectorsSummaryResponseData, error) {
	// 获取所有活跃的代码库
	codebases, err := l.svcCtx.Querier.Codebase.WithContext(l.ctx).
		Where(l.svcCtx.Querier.Codebase.Status.Eq(string(model.CodebaseStatusActive))).
		Find()
	if err != nil {
		return nil, fmt.Errorf("failed to get all codebases: %w", err)
	}

	if len(codebases) == 0 {
		return &types.GetAllVectorsSummaryResponseData{
			TotalCount: 0,
			Items:      []*types.VectorSummaryItem{},
		}, nil
	}

	// 定义超时时间
	timeout := 10 * time.Second

	// 并发获取每个代码库的向量信息
	var wg sync.WaitGroup
	items := make([]*types.VectorSummaryItem, len(codebases))
	errChan := make(chan error, len(codebases))

	for i, codebase := range codebases {
		wg.Add(1)
		go func(idx int, cb *model.Codebase) {
			defer wg.Done()

			ctx := context.WithValue(l.ctx, tracer.Key, tracer.RequestTraceId(int(cb.ID)))

			// 获取向量索引状态（带超时控制）
			timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			var embeddingSummary *types.EmbeddingSummary
			var embeddingIndexTask *model.IndexHistory

			// 获取向量汇总信息
			embeddingSummary, err = l.svcCtx.VectorStore.GetIndexSummary(timeoutCtx, cb.ClientID, cb.Path)
			if err != nil {
				if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
					tracer.WithTrace(ctx).Errorf("embedding summary query timed out after %v for codebase %d", timeout, cb.ID)
				} else {
					tracer.WithTrace(ctx).Errorf("failed to get embedding summary for codebase %d, err:%v", cb.ID, err)
				}
				// 不返回错误，继续处理其他代码库
			}

			// 获取最新的索引任务状态
			embeddingIndexTask, err = l.svcCtx.Querier.IndexHistory.GetLatestTaskHistory(timeoutCtx, cb.ID, types.TaskTypeEmbedding)
			if err != nil {
				if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
					tracer.WithTrace(timeoutCtx).Errorf("embedding index task query timed out after %v for codebase %d", timeout, cb.ID)
				} else {
					tracer.WithTrace(timeoutCtx).Errorf("failed to get latest embedding index task for codebase %d, err:%v", cb.ID, err)
				}
				// 不返回错误，继续处理其他代码库
			}

			// 构建向量汇总项
			item := &types.VectorSummaryItem{
				ClientId:     cb.ClientID,
				CodebasePath: cb.ClientPath,
				CodebaseName: cb.Name,
				TotalFiles:   int(cb.FileCount),
				Embedding: types.EmbeddingSummary{
					Status: types.TaskStatusPending,
				},
				CreatedAt: cb.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt: cb.UpdatedAt.Format("2006-01-02 15:04:05"),
			}

			// 设置向量状态
			if embeddingIndexTask != nil {
				item.Embedding.Status = convertVectorsStatus(embeddingIndexTask.Status)
				item.Embedding.UpdatedAt = embeddingIndexTask.UpdatedAt.Format("2006-01-02 15:04:05")
			} else if embeddingSummary != nil && embeddingSummary.TotalChunks > 0 {
				item.Embedding.Status = types.TaskStatusSuccess
			}

			// 设置向量统计信息
			if embeddingSummary != nil {
				item.Embedding.TotalChunks = embeddingSummary.TotalChunks
				item.Embedding.TotalFiles = embeddingSummary.TotalFiles
			}

			items[idx] = item
		}(i, codebase)
	}

	// 等待所有协程完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误发生
	for err := range errChan {
		if err != nil {
			logx.Errorf("error occurred while processing codebase: %v", err)
		}
	}

	// 过滤掉nil项
	var validItems []*types.VectorSummaryItem
	for _, item := range items {
		if item != nil {
			validItems = append(validItems, item)
		}
	}

	// 获取任务池状态
	runningTasks := l.svcCtx.TaskPool.Running()
	taskCapacity := l.svcCtx.TaskPool.Cap()

	taskPoolState := &types.TaskPoolState{
		RunningTasks: runningTasks,
		TaskCapacity: taskCapacity,
	}

	return &types.GetAllVectorsSummaryResponseData{
		TotalCount:    len(validItems),
		Items:         validItems,
		TaskPoolState: taskPoolState,
	}, nil
}

func convertVectorsStatus(status string) string {
	var embeddingStatus string
	switch status {
	case types.TaskStatusSuccess:
		embeddingStatus = types.TaskStatusSuccess
	case types.TaskStatusRunning:
		embeddingStatus = types.TaskStatusRunning
	case types.TaskStatusPending:
		embeddingStatus = types.TaskStatusPending
	default:
		embeddingStatus = types.TaskStatusFailed
	}
	return embeddingStatus
}
