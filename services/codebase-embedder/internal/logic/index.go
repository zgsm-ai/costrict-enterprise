package logic

import (
	"context"
	"fmt"

	"github.com/zgsm-ai/codebase-indexer/internal/store/vector"
	"github.com/zgsm-ai/codebase-indexer/internal/tracer"

	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type IndexLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IndexLogic {
	return &IndexLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IndexLogic) DeleteIndex(req *types.DeleteIndexRequest) (resp *types.DeleteIndexResponseData, err error) {
	clientId := req.ClientId
	filePaths := req.FilePaths

	ctx := context.WithValue(l.ctx, tracer.Key, clientId)

	// 如果filePaths为空，则删除整个工程的嵌入数据
	if filePaths == "" {
		if err = l.svcCtx.VectorStore.DeleteByCodebase(ctx, clientId, req.CodebasePath); err != nil {
			return nil, fmt.Errorf("failed to delete embedding codebase, err:%w", err)
		}
		return &types.DeleteIndexResponseData{}, nil
	}

	if err = l.svcCtx.VectorStore.DeleteDictionary(ctx, filePaths, vector.Options{ClientId: clientId,
		CodebasePath: req.CodebasePath}); err != nil {
		return nil, fmt.Errorf("failed to delete embedding index, err:%w", err)
	}

	return &types.DeleteIndexResponseData{}, nil
}
