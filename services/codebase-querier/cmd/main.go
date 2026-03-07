package main

import (
	"context"
	"flag"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zgsm-ai/codebase-indexer/internal/config"
	"github.com/zgsm-ai/codebase-indexer/internal/handler"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"net/http"
)

var configFile = flag.String("f", "etc/conf.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	logx.MustSetup(c.Log)
	logx.DisableStat()

	server := rest.MustNewServer(c.RestConf, rest.WithFileServer("/swagger/", http.Dir("api/docs/")))
	defer server.Stop()

	serverCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	svcCtx, err := svc.NewServiceContext(serverCtx, c)
	defer svcCtx.Close()
	if err != nil {
		panic(err)
	}

	handler.RegisterHandlers(server, svcCtx)

	logx.Infof("==>Started server at %s:%d", c.Host, c.Port)
	server.Start()
}
