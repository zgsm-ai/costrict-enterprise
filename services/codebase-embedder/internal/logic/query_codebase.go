package logic

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/zgsm-ai/codebase-indexer/internal/dao/model"
	"github.com/zgsm-ai/codebase-indexer/internal/errs"
	"github.com/zgsm-ai/codebase-indexer/internal/store/vector"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// QueryCodebaseLogic 查询代码库业务逻辑
type QueryCodebaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewQueryCodebaseLogic 创建查询代码库业务逻辑实例
func NewQueryCodebaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCodebaseLogic {
	return &QueryCodebaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryCodebase 查询代码库信息
func (l *QueryCodebaseLogic) QueryCodebase(req *types.CodebaseQueryRequest) (*types.CodebaseQueryResponse, error) {
	// 1. 验证参数
	if err := l.validateRequest(req); err != nil {
		return nil, err
	}

	// 2. 权限验证：检查clientId与codebase的关联关系
	codebaseInfo, err := l.verifyCodebasePermission(req)
	if err != nil {
		return nil, err
	}

	// 3. 创建向量查询存储实例
	vectorStore := vector.NewCodebaseQueryStore(l.svcCtx.VectorStore, l.Logger)

	// 4. 并行查询各种信息，包括详细记录
	var summary *types.CodebaseSummary
	var languageDist []types.LanguageDistribution
	var recentFiles []types.RecentFileInfo
	var indexStats *types.IndexStatistics
	var records []types.CodebaseRecord
	var queryErr error

	// 使用goroutine并行查询提高性能
	done := make(chan bool, 1)
	go func() {
		// 获取汇总信息
		summary, queryErr = vectorStore.QueryCodebaseStats(l.ctx, req.ClientId, codebaseInfo.ClientPath)
		if queryErr != nil {
			done <- true
			return
		}

		// 获取语言分布
		languageDist, queryErr = vectorStore.QueryLanguageDistribution(l.ctx, codebaseInfo.ID)
		if queryErr != nil {
			done <- true
			return
		}

		// 获取最近文件
		recentFiles, queryErr = vectorStore.QueryRecentFiles(l.ctx, codebaseInfo.ID, 10) // 默认返回最近10个文件
		if queryErr != nil {
			done <- true
			return
		}

		// 获取索引统计
		indexStats, queryErr = vectorStore.QueryIndexStats(l.ctx, codebaseInfo.ID)
		if queryErr != nil {
			done <- true
			return
		}

		// 获取详细记录
		records, queryErr = vectorStore.QueryCodebaseRecords(l.ctx, req.ClientId, codebaseInfo.ClientPath)
		done <- true
	}()

	<-done // 等待所有查询完成

	if queryErr != nil {
		l.Errorf("查询代码库信息失败, clientId: %s, codebaseName: %s, error: %v",
			req.ClientId, req.CodebaseName, queryErr)
		return nil, fmt.Errorf("查询代码库信息失败: %w", queryErr)
	}

	// 5. 构建响应
	response := &types.CodebaseQueryResponse{
		CodebaseId:   codebaseInfo.ID,
		CodebaseName: codebaseInfo.Name,
		CodebasePath: codebaseInfo.ClientPath,
		Summary:      summary,
		LanguageDist: languageDist,
		RecentFiles:  recentFiles,
		IndexStats:   indexStats,
		Records:      records, // 添加详细记录
	}

	// 6. 记录查询日志
	l.logQuery(req, response)

	return response, nil
}

// validateRequest 验证请求参数
func (l *QueryCodebaseLogic) validateRequest(req *types.CodebaseQueryRequest) error {
	if req.ClientId == "" {
		return errs.NewMissingParamError("clientId不能为空")
	}
	if req.CodebasePath == "" {
		return errs.NewMissingParamError("codebasePath不能为空")
	}
	if req.CodebaseName == "" {
		return errs.NewMissingParamError("codebaseName不能为空")
	}
	return nil
}

// verifyCodebasePermission 验证代码库权限
func (l *QueryCodebaseLogic) verifyCodebasePermission(req *types.CodebaseQueryRequest) (*model.Codebase, error) {
	// 使用 Querier 查询数据库验证clientId与codebase的关联关系
	codebase, err := l.svcCtx.Querier.Codebase.WithContext(l.ctx).
		Where(l.svcCtx.Querier.Codebase.ClientID.Eq(req.ClientId)).
		Where(l.svcCtx.Querier.Codebase.ClientPath.Eq(req.CodebasePath)).
		First()

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			l.Infof("代码库不存在或无权限访问, clientId: %s, codebaseName: %s, codebasePath: %s",
				req.ClientId, req.CodebaseName, req.CodebasePath)
			return nil, fmt.Errorf("代码库不存在或无权限访问")
		}
		l.Errorf("查询代码库信息失败, error: %v", err)
		return nil, fmt.Errorf("查询代码库信息失败: %w", err)
	}

	// 检查代码库状态
	if codebase.Status != "active" {
		l.Infof("代码库状态不正常, status: %s", codebase.Status)
		return nil, fmt.Errorf("代码库状态不正常，无法查询")
	}

	return codebase, nil
}

// logQuery 记录查询日志
func (l *QueryCodebaseLogic) logQuery(req *types.CodebaseQueryRequest, resp *types.CodebaseQueryResponse) {
	// 记录查询日志，包含关键信息但避免敏感数据
	totalRecords := len(resp.Records)
	if resp.Summary != nil {
		l.Infof("代码库查询成功, clientId: %s, codebaseId: %d, codebaseName: %s, totalChunks: %d, totalRecords: %d",
			req.ClientId, resp.CodebaseId, resp.CodebaseName, resp.Summary.TotalChunks, totalRecords)
	} else {
		l.Infof("代码库查询成功, clientId: %s, codebaseId: %d, codebaseName: %s, totalRecords: %d",
			req.ClientId, resp.CodebaseId, resp.CodebaseName, totalRecords)
	}
}
