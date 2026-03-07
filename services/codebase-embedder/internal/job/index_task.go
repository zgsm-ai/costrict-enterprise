package job

import (
	"context"
	"fmt"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/tracer"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

type IndexTask struct {
	SvcCtx *svc.ServiceContext
	Params *IndexTaskParams
}

type IndexTaskParams struct {
	SyncID       int32  // 同步操作ID
	CodebaseID   int32  // 代码库ID
	CodebasePath string // 代码库路径
	CodebaseName string // 代码库名字
	ClientId     string // 客户端ID
	RequestId    string // 请求ID，用于状态管理
	Files        map[string][]byte
	Metadata     *types.SyncMetadata // 同步元数据
	TotalFiles   int                 // 文件总数
}

func (i *IndexTask) Run(ctx context.Context) (embedTaskOk bool) {
	start := time.Now()
	tracer.WithTrace(ctx).Infof("index task started")

	// 启动嵌入任务
	embedErr := i.buildEmbedding(ctx)
	if embedErr != nil {
		tracer.WithTrace(ctx).Errorf("embedding task failed:%v", embedErr)
	}

	embedTaskOk = embedErr == nil

	tracer.WithTrace(ctx).Infof("index task end, cost %d ms. embedding ok? %t",
		time.Since(start).Milliseconds(), embedTaskOk)
	return
}

func (i *IndexTask) buildEmbedding(ctx context.Context) error {
	start := time.Now()

	// 添加日志来跟踪参数
	tracer.WithTrace(ctx).Infof("DEBUG: index_task - i.Params.Files length: %d", len(i.Params.Files))
	if i.Params.Metadata != nil {
		tracer.WithTrace(ctx).Infof("DEBUG: index_task - i.Params.Metadata.FileList length: %d", len(i.Params.Metadata.FileList))
	}

	embeddingTimeout, embeddingTimeoutCancel := context.WithTimeout(ctx, i.SvcCtx.Config.IndexTask.EmbeddingTask.Timeout)
	defer embeddingTimeoutCancel()
	eProcessor, err := NewEmbeddingProcessor(i.SvcCtx, i.Params)
	if err != nil {
		return fmt.Errorf("failed to create embedding task processor for message: %d, err: %w", i.Params.SyncID, err)
	}
	err = eProcessor.Process(embeddingTimeout)
	if err != nil {
		return fmt.Errorf("embedding task failed, err:%w", err)
	}
	tracer.WithTrace(ctx).Infof("embedding task end successfully, cost %d ms.", time.Since(start).Milliseconds())
	return nil
}
