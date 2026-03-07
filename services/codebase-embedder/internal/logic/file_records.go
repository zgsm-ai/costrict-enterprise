package logic

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zgsm-ai/codebase-indexer/internal/store/vector"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

type FileRecordsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFileRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileRecordsLogic {
	return &FileRecordsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetFileRecords 获取指定文件的详细记录
func (l *FileRecordsLogic) GetFileRecords(req *types.FileRecordsRequest) (*types.FileRecordsResponse, error) {
	// 参数验证
	if req.CodebasePath == "" {
		return nil, fmt.Errorf("代码库路径不能为空")
	}
	if req.FilePath == "" {
		return nil, fmt.Errorf("文件路径不能为空")
	}

	// 创建向量查询存储
	queryStore := vector.NewCodebaseQueryStore(l.svcCtx.VectorStore, l.Logger)
	if queryStore == nil {
		return nil, fmt.Errorf("向量查询服务未初始化")
	}

	// 查询文件记录
	records, err := queryStore.QueryFileRecords(l.ctx, req.ClientId, req.CodebasePath, req.FilePath)
	if err != nil {
		l.Errorf("查询文件记录失败, codebasePath: %s, filePath: %s, error: %v", req.CodebasePath, req.FilePath, err)
		return nil, fmt.Errorf("查询文件记录失败: %w", err)
	}

	// 构造响应
	response := &types.FileRecordsResponse{
		CodebasePath: req.CodebasePath,
		FilePath:     req.FilePath,
		Records:      records,
		TotalCount:   len(records),
	}

	return response, nil
}
