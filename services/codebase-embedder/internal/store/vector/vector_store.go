package vector

import (
	"context"
	"errors"
	"fmt"

	"github.com/zgsm-ai/codebase-indexer/internal/config"
	"github.com/zgsm-ai/codebase-indexer/internal/store/redis"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// Store 向量存储接口
type Store interface {
	DeleteByCodebase(ctx context.Context, clientId string, codebasePath string) error
	GetIndexSummary(ctx context.Context, clientId string, codebasePath string) (*types.EmbeddingSummary, error)
	GetIndexSummaryWithLanguage(ctx context.Context, clientId string, codebasePath string, language string) (*types.EmbeddingSummary, error)
	GetCodebaseRecords(ctx context.Context, clientId string, codebasePath string) ([]*types.CodebaseRecord, error)
	GetFileRecords(ctx context.Context, clientId string, codebasePath string, filePath string) ([]*types.CodebaseRecord, error)
	GetDictionaryRecords(ctx context.Context, clientId string, codebasePath string, dictionary string) ([]*types.CodebaseRecord, error)
	InsertCodeChunks(ctx context.Context, docs []*types.CodeChunk, options Options) error
	UpsertCodeChunks(ctx context.Context, chunks []*types.CodeChunk, options Options) error
	DeleteCodeChunks(ctx context.Context, chunks []*types.CodeChunk, options Options) error
	DeleteDictionary(ctx context.Context, dictionary string, options Options) error
	UpdateCodeChunksPaths(ctx context.Context, updates []*types.CodeChunkPathUpdate, options Options) error
	UpdateCodeChunksDictionary(ctx context.Context, clientId string, codebasePath string, dictionary string, newDictionary string) error
	Query(ctx context.Context, query string, topK int, options Options) ([]*types.SemanticFileItem, error)
	Close()
}

const vectorWeaviate = "weaviate"

type Options struct {
	CodebaseId    int32
	SyncId        int32
	RequestId     string
	CodebasePath  string
	CodebaseName  string
	TotalFiles    int
	ClientId      string
	Authorization string
	Language      string
}

func NewVectorStore(cfg config.VectorStoreConf, embedder Embedder, reranker Reranker) (Store, error) {
	return NewVectorStoreWithStatusManager(cfg, embedder, reranker, nil, "")
}

func NewVectorStoreWithStatusManager(cfg config.VectorStoreConf, embedder Embedder, reranker Reranker, statusManager interface{}, requestId string) (Store, error) {
	var vectorStoreImpl Store
	var err error
	switch cfg.Type {
	case vectorWeaviate:
		if cfg.Weaviate.Endpoint == types.EmptyString {
			return nil, errors.New("vector conf weaviate is required for weaviate type")
		}
		// 类型断言，确保 statusManager 是正确的类型
		var sm *redis.StatusManager
		if statusManager != nil {
			if manager, ok := statusManager.(*redis.StatusManager); ok {
				sm = manager
			}
		}
		vectorStoreImpl, err = NewWithStatusManager(cfg, embedder, reranker, sm, requestId)
	default:
		err = fmt.Errorf("unsupported vector type: %s", cfg.Type)
	}

	if err != nil {
		return nil, err
	}
	return vectorStoreImpl, nil
}
