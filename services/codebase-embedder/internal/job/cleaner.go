package job

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zgsm-ai/codebase-indexer/internal/dao/model"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
)

const cleanLockKey = "codebase_embedder:lock:cleaner"
const lockTimeout = time.Second * 120

type Cleaner struct {
	svcCtx *svc.ServiceContext
	ctx    context.Context
	cron   *cron.Cron
}

func (c *Cleaner) Close() {
	//TODO implement me
	// panic("implement me")
}

func (c *Cleaner) Start() {
	c.cron.Start() // 启动 Cron
	logx.Infof("cleaner job started")
}

func NewCleaner(ctx context.Context, svcCtx *svc.ServiceContext) (Job, error) {
	// go cron
	cr := cron.New() // 创建默认 Cron 实例（支持秒级精度）
	// 添加任务（参数：Cron 表达式, 要执行的函数）
	_, err := cr.AddFunc(svcCtx.Config.Cleaner.Cron, func() {

		expireDays := time.Duration(svcCtx.Config.Cleaner.CodebaseExpireDays) * 24 * time.Hour
		expiredDate := time.Now().Add(-expireDays)

		codebases, err := findExpiredCodebases(ctx, svcCtx, expiredDate)
		if err != nil {
			logx.Errorf("find expired codebase error: %v", err)
			return
		}
		for _, cb := range codebases {
			logx.Infof("start to clean codebase: %s", cb.Path)

			// todo clean vector store
			err = svcCtx.VectorStore.DeleteByCodebase(ctx, cb.ClientID, cb.Path)
			if err != nil {
				logx.Errorf("cleaner drop codebase store %s error: %v", cb.Path, err)
			}

			// todo update db status， 唯一索引的存在(client_id、codebasePath)，给client_id 加个唯一后缀，避免冲突。
			cb.ClientID = cb.ClientID + "@" + uuid.New().String()
			cb.Status = string(model.CodebaseStatusExpired)
			if _, err = svcCtx.Querier.Codebase.WithContext(ctx).
				Where(svcCtx.Querier.Codebase.ID.Eq(cb.ID)).
				Updates(cb); err != nil {
				logx.Errorf("cleaner update codebase %s status expired error: %v", cb.Path, err)
				return
			}
			logx.Infof("cleaner clean codebase successfully: %s", cb.Path)
			// TODO sync_history 表清理。（或者直接给表加个触发器）
			if _, err = svcCtx.Querier.IndexHistory.WithContext(ctx).
				Where(svcCtx.Querier.IndexHistory.CodebaseID.Eq(cb.ID)).
				Delete(); err != nil {
				logx.Errorf("cleaner delete index history for codebase %s error: %v", cb.Path, err)
			}

			// TODO 元数据文件定时清理，每天将昨天的清理掉，避免有些任务索引没构建成功，导致任务一直失败。
		}
		logx.Infof("cleaner clean codebases end, cnt: %d", len(codebases))
	})
	if err != nil {
		return nil, err
	}
	return &Cleaner{
		svcCtx: svcCtx,
		ctx:    ctx,
		cron:   cr,
	}, nil
}

func findExpiredCodebases(ctx context.Context, svcCtx *svc.ServiceContext, expiredDate time.Time) ([]*model.Codebase, error) {
	codebases, err := svcCtx.Querier.Codebase.WithContext(ctx).
		Where(svcCtx.Querier.Codebase.UpdatedAt.Lt(expiredDate)).
		Where(svcCtx.Querier.Codebase.Status.Eq(string(model.CodebaseStatusActive))).
		Find()
	return codebases, err
}
