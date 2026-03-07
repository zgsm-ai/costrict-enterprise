package vector

import (
	"context"
	"strings"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/config"
	"github.com/zgsm-ai/codebase-indexer/internal/store/redis"
	"github.com/zgsm-ai/codebase-indexer/internal/tracer"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// Embedder defines the interface for embedding operations
type Embedder interface {
	// EmbedCodeChunks creates embeddings for multiple code chunks
	EmbedCodeChunks(ctx context.Context, chunks []*types.CodeChunk) ([]*CodeChunkEmbedding, error)
	// EmbedQuery creates an embedding for a single query string
	EmbedQuery(ctx context.Context, query string) ([]float32, error)
}

// CodeChunkEmbedding represents a code chunk with its embedding vector
type CodeChunkEmbedding struct {
	*types.CodeChunk
	Embedding []float32
}

// customEmbedder implements the Embedder interface
type customEmbedder struct {
	config          config.EmbedderConf
	embeddingClient EmbeddingClient
	statusManager   *redis.StatusManager
	requestId       string
	totalFiles      int
}

// NewEmbedder creates a new instance of Embedder
func NewEmbedder(cfg config.EmbedderConf) (Embedder, error) {
	embeddingClient := NewEmbeddingClient(cfg)

	return &customEmbedder{
		embeddingClient: embeddingClient,
		config:          cfg,
	}, nil
}

// NewEmbedderWithStatusManager creates a new instance of Embedder with status manager
func NewEmbedderWithStatusManager(cfg config.EmbedderConf, statusManager *redis.StatusManager, requestId string, totalFiles int) (Embedder, error) {
	embeddingClient := NewEmbeddingClient(cfg)

	return &customEmbedder{
		embeddingClient: embeddingClient,
		config:          cfg,
		statusManager:   statusManager,
		requestId:       requestId,
		totalFiles:      totalFiles,
	}, nil
}

// EmbedCodeChunks implements the Embedder interface
func (e *customEmbedder) EmbedCodeChunks(ctx context.Context, chunks []*types.CodeChunk) ([]*CodeChunkEmbedding, error) {
	if len(chunks) == 0 {
		return []*CodeChunkEmbedding{}, nil
	}

	embeds := make([]*CodeChunkEmbedding, 0, len(chunks))
	batchSize := e.config.BatchSize
	start := time.Now()
	tracer.WithTrace(ctx).Infof("start to embedding %d chunks for codebase:%s, batchSize: %d, requestId:%s", len(chunks), chunks[0].CodebasePath, batchSize, e.requestId)

	// 用于跟踪已处理的文件数量
	processedFiles := 0
	// 用于跟踪已处理的文件路径，避免重复计数
	processedFilePaths := make(map[string]bool)
	// 用于跟踪当前批次的文件，每10个文件更新一批
	currentBatchFiles := make([]string, 0, 10)

	for start := 0; start < len(chunks); start += batchSize {
		end := start + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}

		// 准备当前批次的内容
		batch := make([][]byte, end-start)
		for i := 0; i < end-start; i++ {
			batch[i] = chunks[start+i].Content
			filePath := chunks[start+i].FilePath
			if !processedFilePaths[filePath] {
				tracer.WithTrace(ctx).Infof("execute to %s embedding lens: %d, requestid %s", chunks[start+i].FilePath, len(batch[i]), e.requestId)
			}

		}

		// 执行嵌入
		startTime := time.Now()
		embeddings, err := e.doEmbeddings(ctx, batch)
		duration := time.Since(startTime)
		tracer.WithTrace(ctx).Infof("doEmbeddings execution time: %v", duration)
		if err != nil {
			tracer.WithTrace(ctx).Errorf("e.doEmbeddings(ctx, batch) filed: %v ", err)
			break
		}

		// 将嵌入结果与原始块关联
		for i, em := range embeddings {
			embeds = append(embeds, &CodeChunkEmbedding{
				CodeChunk: chunks[start+i],
				Embedding: em,
			})
			filePath := chunks[start+i].FilePath
			if !processedFilePaths[filePath] {
				processedFilePaths[filePath] = true
				processedFiles++
				// 将文件添加到当前批次
				currentBatchFiles = append(currentBatchFiles, filePath)

				// 每处理10个文件就同步一次进度，并将这批文件状态改为completed
				// 或者处理完所有文件时，如果当前批次有文件也要更新（处理最后一批不满10个的情况）
				shouldUpdate := (processedFiles%10 == 0) || (processedFiles == len(processedFilePaths) && len(currentBatchFiles) > 0)

				if shouldUpdate && e.statusManager != nil && e.requestId != "" {
					// 使用总文件数计算进度，如果总文件数为0则使用已处理的文件路径数量作为分母
					var denominator int
					if e.totalFiles > 0 {
						denominator = e.totalFiles
					} else {
						denominator = len(processedFilePaths)
					}
					progress := int(float64(processedFiles) / float64(denominator) * 100)

					// 复制当前批次文件作为要更新的文件列表
					completedFiles := make([]string, len(currentBatchFiles))
					copy(completedFiles, currentBatchFiles)

					err := e.statusManager.UpdateFileStatus(ctx, e.requestId, func(status *types.FileStatusResponseData) {
						status.Process = "processing"
						status.TotalProgress = progress

						// tracer.WithTrace(ctx).Infof(" requestid %s  ,finish before %v", e.requestId, status.FileList)

						// 将这批文件的状态添加到FileList中
						for _, filePath := range completedFiles {
							// 检查文件是否已在FileList中
							// found := false
							for i, item := range status.FileList {
								if item.Path == filePath {
									// 更新现有文件状态
									status.FileList[i].Status = "completed"
									// found = true
									break
								}
							}
							// 如果文件不在FileList中，则添加新项
							// if !found {
							// 	tracer.WithTrace(ctx).Infof("not found in filePath: %v  ,fileList %v", filePath, status.FileList)
							// 	status.FileList = append(status.FileList, types.FileStatusItem{
							// 		Path:    filePath,
							// 		Status:  "completed",
							// 		Operate: "add", // 默认操作类型为add
							// 	})
							// }
						}

						// tracer.WithTrace(ctx).Infof(" requestid %s  ,finish %v", e.requestId, status.FileList)

					})
					if err != nil {
						tracer.WithTrace(ctx).Errorf("failed to update progress: %v", err)
					} else {
						tracer.WithTrace(ctx).Infof("updated progress: %d%% (%d/%d files), marked %d files as completed", progress, processedFiles, len(processedFilePaths), len(completedFiles))
					}

					// 清空当前批次，为下一批做准备
					currentBatchFiles = make([]string, 0, 10)
				}
			}
		}
	}

	// 最终更新一次进度
	if e.statusManager != nil && e.requestId != "" {
		progress := 100
		// 计算最终的分母用于日志显示
		var denominator int
		if e.totalFiles > 0 {
			denominator = e.totalFiles
		} else {
			denominator = len(processedFilePaths)
		}
		err := e.statusManager.UpdateFileStatus(ctx, e.requestId, func(status *types.FileStatusResponseData) {
			status.Process = "complete"
			status.TotalProgress = progress
		})
		if err != nil {
			tracer.WithTrace(ctx).Errorf("failed to update final progress: %v", err)
		} else {
			tracer.WithTrace(ctx).Infof("updated final progress: %d%% (%d/%d files)", progress, processedFiles, denominator)
		}
	}

	tracer.WithTrace(ctx).Infof("embedding %d chunks for codebase:%s successfully, cost %d ms", len(chunks),
		chunks[0].CodebasePath, time.Since(start).Milliseconds())

	return embeds, nil
}

// EmbedQuery implements the Embedder interface
func (e *customEmbedder) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	if e.config.StripNewLines {
		query = strings.ReplaceAll(query, "\n", " ")
	}
	tracer.WithTrace(ctx).Info("start to embed query")
	vectors, err := e.doEmbeddings(ctx, [][]byte{[]byte(query)})
	if err != nil {

		return nil, err
	}
	if len(vectors) == 0 {
		return nil, ErrEmptyResponse
	}
	tracer.WithTrace(ctx).Info("embed query successfully")
	return vectors[0], nil
}

// doEmbeddings performs the actual embedding operation
func (e *customEmbedder) doEmbeddings(ctx context.Context, textsByte [][]byte) ([][]float32, error) {
	texts := make([]string, len(textsByte))
	for i, b := range textsByte {
		texts[i] = string(b)
	}

	embeddings, err := e.embeddingClient.CreateEmbeddings(ctx, texts, e.config.Model)
	if err != nil {
		for _, text := range texts {
			tracer.WithTrace(ctx).Errorf("embed query file len %d failed, err: %v", len(text), err)
		}

		return nil, err
	}

	vectors := make([][]float32, len(textsByte))
	for i, embedding := range embeddings {
		transferredVector := make([]float32, 0, 768) // 768维
		for _, v := range embedding {
			transferredVector = append(transferredVector, float32(v))
		}
		vectors[i] = transferredVector
	}
	return vectors, nil
}
