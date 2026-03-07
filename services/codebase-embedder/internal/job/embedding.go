package job

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/errs"
	"github.com/zgsm-ai/codebase-indexer/internal/parser"
	"github.com/zgsm-ai/codebase-indexer/internal/store/vector"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/tracer"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

type embeddingProcessor struct {
	baseProcessor
}

func NewEmbeddingProcessor(
	svcCtx *svc.ServiceContext,
	msg *IndexTaskParams,
) (Processor, error) {
	return &embeddingProcessor{
		baseProcessor: baseProcessor{
			svcCtx: svcCtx,
			params: msg,
		},
	}, nil
}

type fileProcessResult struct {
	chunks []*types.CodeChunk
	err    error
	path   string
}

func (t *embeddingProcessor) Process(ctx context.Context) error {
	tracer.WithTrace(ctx).Infof("start to execute embedding task, codebase: %s RequestId %s", t.params.CodebaseName, t.params.RequestId)
	start := time.Now()

	err := func(t *embeddingProcessor) error {
		if err := t.initTaskHistory(ctx, types.TaskTypeEmbedding); err != nil {
			return err
		}

		t.totalFileCnt = int32(len(t.params.Files))

		// 添加日志来跟踪文件数量
		tracer.WithTrace(ctx).Infof("DEBUG: embedding task - totalFileCnt: %d", t.totalFileCnt)
		tracer.WithTrace(ctx).Infof("DEBUG: embedding task - t.params.Files length: %d", len(t.params.Files))

		var (
			addChunks        = make([]*types.CodeChunk, 0, t.totalFileCnt)
			deleteFilePaths  = make(map[string]struct{})
			unsupportedFiles = make([]string, 0) // 收集不支持的文件路径
			mu               sync.Mutex          // 保护 addChunks 和 unsupportedFiles
		)

		// 处理单个文件的函数
		processFile := func(path string, content []byte) error {
			select {
			case <-ctx.Done():
				return errs.RunTimeout
			default:
				chunks, err := t.splitFile(&types.SourceFile{Path: path, Content: content})
				if err != nil {
					mu.Lock()
					unsupportedFiles = append(unsupportedFiles, path)
					mu.Unlock()

					if parser.IsNotSupportedFileError(err) {

						atomic.AddInt32(&t.ignoreFileCnt, 1)
						return nil
					}
					atomic.AddInt32(&t.failedFileCnt, 1)
					return err
				}
				mu.Lock()

				if len(chunks) <= 0 {
					unsupportedFiles = append(unsupportedFiles, path)
				}

				addChunks = append(addChunks, chunks...)
				mu.Unlock()

				atomic.AddInt32(&t.successFileCnt, 1)

			}
			return nil
		}

		// 使用基础结构的并发处理方法
		if err := t.processFilesConcurrently(ctx, processFile, t.svcCtx.Config.IndexTask.EmbeddingTask.MaxConcurrency); err != nil {

			if len(unsupportedFiles) > 0 {
				tracer.WithTrace(ctx).Infof("updating %d unsupported files status", len(unsupportedFiles))
				err := t.svcCtx.StatusManager.UpdateFileStatus(ctx, t.params.RequestId,
					func(status *types.FileStatusResponseData) {
						status.Process = "completed"
						status.TotalProgress = 100
						for _, filePath := range unsupportedFiles {
							for i, item := range status.FileList {
								if item.Path == filePath {
									status.FileList[i].Status = "unsupported"
									tracer.WithTrace(ctx).Infof("marked file as unsupported: %s", filePath)
									break
								}
							}
						}
					})
				if err != nil {
					tracer.WithTrace(ctx).Errorf("failed to update unsupported files status: %v", err)
				}
			}
			return err
		}

		// 统一更新不支持的文件状态
		if len(unsupportedFiles) > 0 {
			tracer.WithTrace(ctx).Infof("updating %d unsupported files status", len(unsupportedFiles))
			// 更新不支持的文件状态
			err := t.svcCtx.StatusManager.UpdateFileStatus(ctx, t.params.RequestId,
				func(status *types.FileStatusResponseData) {
					status.Process = "completed"
					status.TotalProgress = 100
					for _, filePath := range unsupportedFiles {
						for i, item := range status.FileList {
							if item.Path == filePath {
								status.FileList[i].Status = "unsupported"
								tracer.WithTrace(ctx).Infof("marked file as unsupported: %s", filePath)
								break
							}
						}
					}
				})
			if err != nil {
				tracer.WithTrace(ctx).Errorf("failed to update unsupported files status: %v", err)
			}
		}

		// 打印不支持文件个数
		tracer.WithTrace(ctx).Infof("embedding splitFile successfully, cost: %d ms, total: %d,success %d ,  unsupported: %d",
			time.Since(start).Milliseconds(), t.totalFileCnt, t.successFileCnt, t.ignoreFileCnt)

		var saveErrs []error
		// 先删除，再写入
		if len(deleteFilePaths) > 0 {
			var deleteChunks []*types.CodeChunk
			for path := range deleteFilePaths {
				deleteChunks = append(deleteChunks, &types.CodeChunk{
					CodebaseId:   t.params.CodebaseID,
					CodebasePath: t.params.CodebasePath,
					CodebaseName: t.params.CodebaseName,
					FilePath:     path,
				})
			}
			err := t.svcCtx.VectorStore.DeleteCodeChunks(ctx, deleteChunks, vector.Options{
				ClientId:     t.params.ClientId,
				CodebaseId:   t.params.CodebaseID,
				CodebasePath: t.params.CodebasePath,
				CodebaseName: t.params.CodebaseName,
				SyncId:       t.params.SyncID,
				RequestId:    t.params.RequestId,
			})
			if err != nil {
				tracer.WithTrace(ctx).Errorf("embedding task delete code chunks failed: %v", err)
				t.failedFileCnt += int32(len(deleteFilePaths))
				saveErrs = append(saveErrs, err)
			}
		}

		// 批量处理结果
		if len(addChunks) > 0 {
			err := t.svcCtx.VectorStore.UpsertCodeChunks(ctx, addChunks, vector.Options{
				ClientId:     t.params.ClientId,
				CodebaseId:   t.params.CodebaseID,
				CodebasePath: t.params.CodebasePath,
				CodebaseName: t.params.CodebaseName,
				SyncId:       t.params.SyncID,
				RequestId:    t.params.RequestId,
				TotalFiles:   t.params.TotalFiles,
			})
			if err != nil {
				tracer.WithTrace(ctx).Errorf("embedding task upsert code chunks failed: %v", err)
				t.failedFileCnt += t.successFileCnt
				t.successFileCnt = 0
				saveErrs = append(saveErrs, err)
			}
		}

		// 更新最终状态
		t.svcCtx.StatusManager.UpdateFileStatus(ctx, t.params.RequestId,
			func(status *types.FileStatusResponseData) {
				status.Process = "completed"
				status.TotalProgress = 100
			})

		if len(saveErrs) > 0 {
			return errors.Join(saveErrs...)
		}
		// update task status
		if err := t.updateTaskSuccess(ctx); err != nil {
			tracer.WithTrace(ctx).Errorf("embedding task update status success error:%v", err)
		}

		return nil
	}(t)

	if t.handleIfTaskFailed(ctx, err) {
		return fmt.Errorf("embedding task failed to update status, err:%v", err)
	}

	tracer.WithTrace(ctx).Infof("embedding task end successfully, cost: %d ms, total: %d, success: %d, failed: %d, unsupported: %d",
		time.Since(start).Milliseconds(), t.totalFileCnt, t.successFileCnt, t.failedFileCnt, t.ignoreFileCnt)
	return nil
}

func (t *embeddingProcessor) splitFile(file *types.SourceFile) ([]*types.CodeChunk, error) {
	// 切分文件
	return t.svcCtx.CodeSplitter.Split(&types.SourceFile{
		CodebaseId:   t.params.CodebaseID,
		CodebasePath: t.params.CodebasePath,
		CodebaseName: t.params.CodebaseName,
		Path:         file.Path,
		Content:      file.Content,
	})
}
