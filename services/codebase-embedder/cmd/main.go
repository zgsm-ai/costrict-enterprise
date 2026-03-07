package main

import (
	"context"
	"flag"
	"net/http"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zgsm-ai/codebase-indexer/internal/config"
	"github.com/zgsm-ai/codebase-indexer/internal/handler"
	"github.com/zgsm-ai/codebase-indexer/internal/job"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/migrations"
)

var configFile = flag.String("f", "etc/conf.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config

	// 添加调试日志来验证配置加载
	logx.Infof("尝试加载配置文件: %s", *configFile)

	// 验证配置加载前的状态
	logx.Infof("配置加载前，Validation 字段零值: %+v", c.Validation)

	// 尝试加载配置，添加错误处理
	err := conf.Load(*configFile, &c, conf.UseEnv())
	if err != nil {
		logx.Errorf("配置加载失败: %v", err)
		panic(err)
	}
	logx.Infof("配置文件加载成功，Validation 配置: %+v", c.Validation)
	logx.Infof("Validation.Enabled: %v", c.Validation.Enabled)
	logx.Infof("Validation.CheckContent: %v", c.Validation.CheckContent)

	logx.MustSetup(c.Log)
	logx.DisableStat()
	if err := migrations.AutoMigrate(c.Database); err != nil {
		panic(err)
	}

	server := rest.MustNewServer(c.RestConf, rest.WithFileServer("/swagger/", http.Dir("api/docs/")))
	defer server.Stop()

	serverCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	svcCtx, err := svc.NewServiceContext(serverCtx, c)
	defer svcCtx.Close()
	if err != nil {
		panic(err)
	}

	// 在程序启动时重置所有pending和processing任务为failed状态
	logx.Infof("Resetting all pending and processing tasks to failed...")
	if err := svcCtx.StatusManager.ResetPendingAndProcessingTasksToFailed(serverCtx); err != nil {
		logx.Errorf("Failed to reset pending and processing tasks to failed: %v", err)
		// 不应该因为重置失败而阻止程序启动
	} else {
		logx.Infof("Successfully reset all pending and processing tasks to failed")
	}

	jobScheduler, err := job.NewScheduler(serverCtx, svcCtx)
	if err != nil {
		panic(err)
	}
	jobScheduler.Schedule()
	defer jobScheduler.Close()

	handler.RegisterHandlers(server, svcCtx)

	logx.Infof("==>Started server at %s:%d", c.Host, c.Port)
	server.Start()
}
