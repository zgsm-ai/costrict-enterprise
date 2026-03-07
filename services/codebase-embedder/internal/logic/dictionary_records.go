package logic

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zgsm-ai/codebase-indexer/internal/store/vector"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

type DictionaryRecordsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDictionaryRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictionaryRecordsLogic {
	return &DictionaryRecordsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetDictionaryRecords 获取指定目录的详细记录
func (l *DictionaryRecordsLogic) GetDictionaryRecords(req *types.DictionaryRecordsRequest) (*types.DictionaryRecordsResponse, error) {
	// 参数验证
	if req.CodebasePath == "" {
		return nil, fmt.Errorf("代码库路径不能为空")
	}
	if req.Dictionary == "" {
		return nil, fmt.Errorf("目录路径不能为空")
	}

	// 创建向量查询存储
	queryStore := vector.NewCodebaseQueryStore(l.svcCtx.VectorStore, l.Logger)
	if queryStore == nil {
		return nil, fmt.Errorf("向量查询服务未初始化")
	}

	// 查询目录记录
	records, err := queryStore.QueryDictionaryRecords(l.ctx, req.ClientId, req.CodebasePath, req.Dictionary)
	if err != nil {
		l.Errorf("查询目录记录失败, codebasePath: %s, dictionary: %s, error: %v", req.CodebasePath, req.Dictionary, err)
		return nil, fmt.Errorf("查询目录记录失败: %w", err)
	}

	// 构造响应
	response := &types.DictionaryRecordsResponse{
		ClientId:     req.ClientId,
		CodebasePath: req.CodebasePath,
		Dictionary:   req.Dictionary,
		Records:      records,
		TotalCount:   len(records),
	}

	return response, nil
}
