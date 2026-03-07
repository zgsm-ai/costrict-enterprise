package logic

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/dao/model"
	"github.com/zgsm-ai/codebase-indexer/internal/tracer"

	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CombinedSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCombinedSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CombinedSummaryLogic {
	return &CombinedSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CombinedSummaryLogic) CombinedSummary(req *types.CombinedSummaryRequest, authorization string) (*types.CombinedSummaryResponseData, error) {
	var (
		wg                 sync.WaitGroup
		embeddingSummary   *types.EmbeddingSummary
		embeddingIndexTask *model.IndexHistory
		embeddingErr       error
	)

	ctx := context.WithValue(l.ctx, tracer.Key, req.ClientId)

	// 定义超时时间
	timeout := 5 * time.Second

	// 获取向量索引状态（带超时控制）
	wg.Add(1)
	go func() {
		defer wg.Done()
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel() // 避免资源泄漏

		var err error
		embeddingSummary, err = l.svcCtx.VectorStore.GetIndexSummary(timeoutCtx, req.ClientId, req.CodebasePath)
		if err != nil {
			if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
				tracer.WithTrace(ctx).Errorf("embedding summary query timed out after %v", timeout)
				embeddingErr = errors.New("embedding summary query timed out")
			} else {
				tracer.WithTrace(ctx).Errorf("failed to get embedding summary, err:%v", err)
				embeddingErr = err
			}
			return
		}
	}()

	// 等待所有协程完成
	wg.Wait()

	// 检查是否有错误发生
	if embeddingErr != nil {
		return nil, embeddingErr
	}

	resp := &types.CombinedSummaryResponseData{
		TotalFiles: 0,
		Embedding: types.EmbeddingSummary{
			Status: types.TaskStatusPending,
		},
	}

	if embeddingIndexTask != nil {
		resp.Embedding.Status = convertCombinedStatus(embeddingIndexTask.Status)
		resp.Embedding.UpdatedAt = embeddingIndexTask.UpdatedAt.Format("2006-01-02 15:04:05")
	} else if embeddingSummary.TotalChunks > 0 {
		resp.Embedding.Status = types.TaskStatusSuccess
	}

	if embeddingSummary != nil {
		resp.TotalFiles = embeddingSummary.TotalFiles
		resp.Embedding.TotalChunks = embeddingSummary.TotalChunks
		resp.Embedding.TotalFiles = embeddingSummary.TotalFiles
	}

	return resp, nil
}

func convertCombinedStatus(status string) string {
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
