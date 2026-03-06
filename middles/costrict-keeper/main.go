// @title Costrict Keeper API
// @version 1.0
// @description This is the API server for Costrict Keeper
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8999
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	_ "costrict-keeper/cmd"
	"costrict-keeper/cmd/root"
	"costrict-keeper/internal/config"
	"costrict-keeper/internal/logger"
	"os"
)

func main() {
	// 检查是否是服务器模式
	isServerMode := len(os.Args) > 1 && os.Args[1] == "server"
	config.LoadConfig(true)
	cfg := config.App()
	logger.InitLogger(cfg.Log.Path, cfg.Log.Level, isServerMode, cfg.Log.MaxSize, cfg.Log.Backup)

	if err := root.RootCmd.Execute(); err != nil {
		logger.Fatal(err)
	}
	os.Exit(0)
}
