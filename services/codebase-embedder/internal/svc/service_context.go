package svc

import (
	"context"

	"github.com/panjf2000/ants/v2"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zgsm-ai/codebase-indexer/internal/config"
	"github.com/zgsm-ai/codebase-indexer/internal/dao/query"
	"github.com/zgsm-ai/codebase-indexer/internal/embedding"
	"github.com/zgsm-ai/codebase-indexer/internal/store/database"
	redisstore "github.com/zgsm-ai/codebase-indexer/internal/store/redis"
	"github.com/zgsm-ai/codebase-indexer/internal/store/vector"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config        config.Config
	db            *gorm.DB
	Querier       *query.Query
	Embedder      vector.Embedder
	VectorStore   vector.Store
	CodeSplitter  *embedding.CodeSplitter
	StatusManager *redisstore.StatusManager
	redisClient   *redis.Client // 保存Redis客户端引用以便关闭
	serverContext context.Context
	TaskPool      *ants.Pool
}

// Close closes the shared Redis client and database connection
func (s *ServiceContext) Close() {
	var errs []error
	if s.redisClient != nil {
		if err := s.redisClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.db != nil {
		if err := database.CloseDB(s.db); err != nil {
			errs = append(errs, err)
		}
	}
	if s.TaskPool != nil {
		s.TaskPool.Release()
	}
	if len(errs) > 0 {
		logx.Errorf("service_context close err:%v", errs)
	} else {
		logx.Infof("service_context close successfully.")
	}
}

func NewServiceContext(ctx context.Context, c config.Config) (*ServiceContext, error) {
	var err error
	svcCtx := &ServiceContext{
		Config:        c,
		serverContext: ctx,
	}

	// 初始化数据库连接
	db, err := database.New(c.Database)
	if err != nil {
		return nil, err
	}
	svcCtx.db = db

	querier := query.Use(db)
	svcCtx.Querier = querier

	// 创建Redis客户端
	client, err := redisstore.NewRedisClient(c.Redis)
	if err != nil {
		return nil, err
	}
	svcCtx.redisClient = client

	embedder, err := vector.NewEmbedder(c.VectorStore.Embedder)
	if err != nil {
		return nil, err
	}
	reranker := vector.NewReranker(c.VectorStore.Reranker)

	splitter, err := embedding.NewCodeSplitter(embedding.SplitOptions{
		MaxTokensPerChunk:          c.IndexTask.EmbeddingTask.MaxTokensPerChunk,
		SlidingWindowOverlapTokens: c.IndexTask.EmbeddingTask.OverlapTokens,
		EnableMarkdownParsing:      c.IndexTask.EmbeddingTask.EnableMarkdownParsing,
		EnableOpenAPIParsing:       c.IndexTask.EmbeddingTask.EnableOpenAPIParsing,
	})
	if err != nil {
		return nil, err
	}

	// 初始化协程池
	taskPool, err := ants.NewPool(svcCtx.Config.IndexTask.PoolSize, ants.WithOptions(
		ants.Options{
			MaxBlockingTasks: svcCtx.Config.IndexTask.QueueSize, // max queue tasks, if queue is full, will block
			Nonblocking:      false,
		},
	))
	if err != nil {
		return nil, err
	}
	svcCtx.TaskPool = taskPool

	svcCtx.Embedder = embedder
	svcCtx.CodeSplitter = splitter
	// 状态管理器 - 使用配置中的默认过期时间
	svcCtx.StatusManager = redisstore.NewStatusManagerWithExpiration(client, c.Redis.DefaultExpiration)

	// 向量知识库
	vectorStore, err := vector.NewVectorStoreWithStatusManager(c.VectorStore, embedder, reranker, svcCtx.StatusManager, "")
	if err != nil {
		return nil, err
	}
	svcCtx.VectorStore = vectorStore

	return svcCtx, err
}
