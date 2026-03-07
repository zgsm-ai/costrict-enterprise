package logic

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

type UpdateEmbeddingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateEmbeddingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateEmbeddingLogic {
	return &UpdateEmbeddingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateEmbeddingLogic) UpdateEmbeddingPath(req *types.UpdateEmbeddingPathRequest) (resp *types.UpdateEmbeddingPathResponseData, err error) {
	codebasePath := req.CodebasePath
	oldPath := req.OldPath
	newPath := req.NewPath

	var modifiedFiles []string

	// 处理目录情况，使用 UpdateCodeChunksDictionary 接口
	err = l.svcCtx.VectorStore.UpdateCodeChunksDictionary(l.ctx, req.ClientId, codebasePath, oldPath, newPath)
	if err != nil {
		return nil, fmt.Errorf("failed to update directory paths: %w", err)
	}

	// 获取更新后的记录以返回修改的文件列表
	records, err := l.svcCtx.VectorStore.GetDictionaryRecords(l.ctx, req.ClientId, codebasePath, newPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated dictionary records: %w", err)
	}

	for _, record := range records {
		modifiedFiles = append(modifiedFiles, record.FilePath)
	}

	return &types.UpdateEmbeddingPathResponseData{
		ModifiedFiles: modifiedFiles,
		TotalFiles:    len(modifiedFiles),
	}, nil
}
