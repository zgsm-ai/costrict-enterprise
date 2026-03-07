package job

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/tracer"

	"github.com/panjf2000/ants/v2"
	"github.com/zgsm-ai/codebase-indexer/internal/errs"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	// "github.com/zgsm-ai/codebase-indexer/internal/types"
)

// baseProcessor 包含所有处理器共有的字段和方法
type baseProcessor struct {
	svcCtx         *svc.ServiceContext
	params         *IndexTaskParams
	taskHistoryId  int32
	totalFileCnt   int32
	successFileCnt int32
	failedFileCnt  int32
	ignoreFileCnt  int32
}

// initTaskHistory 初始化任务历史记录
func (p *baseProcessor) initTaskHistory(ctx context.Context, taskType string) error {
	// taskHistory := &model.IndexHistory{
	// 	SyncID:       p.params.SyncID,
	// 	CodebaseID:   p.params.CodebaseID,
	// 	CodebasePath: p.params.CodebasePath,
	// 	TaskType:     taskType,
	// 	Status:       types.TaskStatusPending,
	// 	StartTime:    utils.CurrentTime(),
	// }
	// if err := p.svcCtx.Querier.IndexHistory.WithContext(ctx).Save(taskHistory); err != nil {
	// 	tracer.WithTrace(ctx).Errorf("insert task history failed: %v, data:%v", err, taskHistory)
	// 	return errs.InsertDatabaseFailed
	// }
	// p.taskHistoryId = taskHistory.ID
	return nil
}

// updateTaskSuccess 更新任务状态为成功
func (p *baseProcessor) updateTaskSuccess(ctx context.Context) error {
	// progress := float64(1)
	// m := &model.IndexHistory{
	// 	ID:                p.taskHistoryId,
	// 	Status:            types.TaskStatusSuccess,
	// 	Progress:          &progress,
	// 	EndTime:           utils.CurrentTime(),
	// 	TotalFileCount:    &p.totalFileCnt,
	// 	TotalSuccessCount: &p.successFileCnt,
	// 	TotalFailCount:    &p.failedFileCnt,
	// 	TotalIgnoreCount:  &p.ignoreFileCnt,
	// }

	// res, err := p.svcCtx.Querier.IndexHistory.WithContext(ctx).
	// 	Where(p.svcCtx.Querier.IndexHistory.ID.Eq(m.ID)).
	// 	Updates(m)
	// if err != nil {
	// 	tracer.WithTrace(ctx).Errorf("update task history %d failed: %v, model:%v", p.params.CodebaseID, err, m)
	// 	return fmt.Errorf("upate task success failed: %w", err)
	// }
	// if res.RowsAffected == 0 {
	// 	tracer.WithTrace(ctx).Errorf("update task history %d failed: %v, model:%v", p.params.CodebaseID, err, m)
	// 	return fmt.Errorf("upate task success failed, codebaseId %d not found in database", p.params.CodebaseID)
	// }
	// if res.Error != nil {
	// 	tracer.WithTrace(ctx).Errorf("update task history %d failed: %v, model:%v", p.params.CodebaseID, err, m)
	// 	return fmt.Errorf("upate task success failed: %w", res.Error)
	// }
	return nil
}

// handleIfTaskFailed 处理任务失败情况
func (p *baseProcessor) handleIfTaskFailed(ctx context.Context, err error) bool {
	return true
	// if err != nil {
	// 	tracer.WithTrace(ctx).Errorf("index task failed, err: %v", err)
	// 	if errors.Is(err, errs.InsertDatabaseFailed) {
	// 		return true
	// 	}
	// 	status := types.TaskStatusFailed
	// 	if errors.Is(err, errs.RunTimeout) {
	// 		status = types.TaskStatusTimeout
	// 	}
	// 	_, err = p.svcCtx.Querier.IndexHistory.WithContext(ctx).
	// 		Where(p.svcCtx.Querier.IndexHistory.ID.Eq(p.taskHistoryId)).
	// 		UpdateColumnSimple(p.svcCtx.Querier.IndexHistory.Status.Value(status),
	// 			p.svcCtx.Querier.IndexHistory.ErrorMessage.Value(err.Error()))
	// 	if err != nil {
	// 		tracer.WithTrace(ctx).Errorf("update task history %d failed: %v", p.params.CodebaseID, err)
	// 	}

	// 	return true
	// }
	// return false
}

// processFilesConcurrently 并发处理文件
func (p *baseProcessor) processFilesConcurrently(
	ctx context.Context,
	processFunc func(path string, content []byte) error,
	maxConcurrency int,
) error {
	if maxConcurrency <= 0 {
		maxConcurrency = 10 // 默认值
	}

	pool, err := ants.NewPool(maxConcurrency)
	if err != nil {
		return fmt.Errorf("create ants pool failed: %w", err)
	}
	defer pool.Release()

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		runErrs []error
		done    = make(chan struct{})
		start   = time.Now()
	)

	totalFiles := len(p.params.Files)

	// 添加日志来验证文件数量
	tracer.WithTrace(ctx).Infof("DEBUG: totalFiles count: %d", totalFiles)
	if totalFiles == 0 {
		tracer.WithTrace(ctx).Errorf("DEBUG: No files to process - this will cause divide by zero error")
		return nil // 或者返回一个特定的错误
	}

	// 提交任务到工作池
	for path, content := range p.params.Files {
		select {
		case <-ctx.Done():
			duration := time.Since(start)
			var avgTime string
			if totalFiles > 0 {
				avgTime = (duration / time.Duration(totalFiles)).Round(time.Microsecond).String()
			} else {
				avgTime = "N/A (no files processed)"
			}
			tracer.WithTrace(ctx).Infof("processed %d files in %v (avg: %s/file)", totalFiles, duration.Round(time.Millisecond), avgTime)
			return errs.RunTimeout
		default:
			wg.Add(1)
			if err := pool.Submit(func() {
				defer wg.Done()
				if err := processFunc(path, content); err != nil {
					mu.Lock()
					runErrs = append(runErrs, err)
					mu.Unlock()
				}
			}); err != nil {
				wg.Done()
				mu.Lock()
				runErrs = append(runErrs, fmt.Errorf("submit task failed: %w", err))
				mu.Unlock()
			}
		}
	}

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(done)
	}()

	// 等待任务完成或上下文取消
	select {
	case <-ctx.Done():
		duration := time.Since(start)
		var avgTime string
		if totalFiles > 0 {
			avgTime = (duration / time.Duration(totalFiles)).Round(time.Microsecond).String()
		} else {
			avgTime = "N/A (no files processed)"
		}
		tracer.WithTrace(ctx).Infof("processed %d files in %v (avg: %s/file)", totalFiles, duration.Round(time.Millisecond), avgTime)
		return errs.RunTimeout
	case <-done:
		duration := time.Since(start)
		var avgTime string
		if totalFiles > 0 {
			avgTime = (duration / time.Duration(totalFiles)).Round(time.Microsecond).String()
		} else {
			avgTime = "N/A (no files processed)"
		}
		tracer.WithTrace(ctx).Infof("processed %d files in %v (avg: %s/file)", totalFiles, duration.Round(time.Millisecond), avgTime)
		if len(runErrs) > 0 {
			if len(runErrs) > 10 {
				return fmt.Errorf("process files failed (showing last 10 errors): %w", errors.Join(runErrs[len(runErrs)-10:]...))
			}
			return fmt.Errorf("process files failed: %w", errors.Join(runErrs...))
		}
		return nil
	}
}
