package logic

import (
	"context"

	"github.com/zgsm-ai/codebase-indexer/internal/errs"
	"github.com/zgsm-ai/codebase-indexer/internal/store/vector"
	"github.com/zgsm-ai/codebase-indexer/internal/tracer"
	"github.com/zgsm-ai/codebase-indexer/pkg/utils"

	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	documentMinPositive = 1
	documentDefaultTopK = 5
	documentParamQuery  = "query"
)

type DocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDocumentSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DocumentLogic {
	return &DocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DocumentLogic) DocumentSearch(req *types.DocumentSearchRequest, authorization string) (resp *types.DocumentSearchResponseData, err error) {
	topK := req.TopK
	if topK < documentMinPositive {
		topK = documentDefaultTopK
	}
	if utils.IsBlank(req.Query) {
		return nil, errs.NewInvalidParamErr(documentParamQuery, req.Query)
	}

	// 预处理查询字符串
	req.Query, err = l.preprocessQuery(req.Query)
	if err != nil {
		return nil, err
	}

	ctx := context.WithValue(l.ctx, tracer.Key, req.ClientId)

	documents, err := l.svcCtx.VectorStore.Query(ctx, req.Query, topK,
		vector.Options{
			CodebaseId:    0,
			ClientId:      req.ClientId,
			CodebasePath:  req.CodebasePath,
			CodebaseName:  "",
			Authorization: authorization,
			Language:      "doc",
		})
	if err != nil {
		return nil, err
	}

	// 分数过滤
	scoreThreshold := req.ScoreThreshold
	filteredDocuments := make([]*types.SemanticFileItem, 0, len(documents))
	for _, doc := range documents {
		if doc.Score >= scoreThreshold {
			filteredDocuments = append(filteredDocuments, doc)
		}
	}

	return &types.DocumentSearchResponseData{
		List: filteredDocuments,
	}, nil
}

// preprocessQuery 执行自定义查询预处理逻辑
func (l *DocumentLogic) preprocessQuery(query string) (string, error) {
	// TODO: 实现自定义预处理逻辑
	// 例如: 去除特殊字符、敏感词过滤等
	return query, nil
}
