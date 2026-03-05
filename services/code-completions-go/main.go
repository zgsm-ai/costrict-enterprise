// @title Code Completions API
// @version 1.0
// @description This is a code completion service API.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	_ "code-completion/docs"
	"code-completion/pkg/config"
	"code-completion/pkg/logger"
	_ "code-completion/pkg/logger"
	"code-completion/pkg/model"
	"code-completion/pkg/stream_controller"
	"code-completion/server"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var SoftwareVer = ""
var BuildTime = ""
var BuildTag = ""
var BuildCommitId = ""

/**
 * 打印应用程序版本信息
 * @description
 * - 输出软件版本号
 * - 显示构建时间
 * - 打印构建标签
 * - 显示构建提交ID
 * - 用于启动时显示版本信息
 * @example
 * PrintVersions()
 * // 输出:
 * // Version 1.0.0
 * // Build Time: 2023-01-01T12:00:00Z
 * // Build Tag: latest
 * // Build Commit ID: abc1234
 */
func PrintVersions() {
	fmt.Printf("Version %s\n", SoftwareVer)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Build Tag: %s\n", BuildTag)
	fmt.Printf("Build Commit ID: %s\n", BuildCommitId)
}

func main() {
	PrintVersions()
	// 初始化时区设置，使程序能够识别容器的TZ环境变量
	initTimeZone()

	// 解析命令行参数
	var (
		port = flag.String("port", "8080", "服务器端口")
		mode = flag.String("mode", "release", "运行模式 (debug/release)")
	)
	flag.Parse()

	// 设置Gin运行模式
	if *mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	logger.SetMode(*mode)
	defer logger.Sync()

	initModels()
	initStreamController()

	// 创建路由
	r := server.SetupRouter()

	// 创建服务器
	addr := ":" + *port
	srv := server.NewServer(addr, r)

	// 启动服务器
	if err := srv.Start(); err != nil {
		logger.Fatal("服务器运行失败", zap.Error(err))
		os.Exit(1)
	}
}

/**
 * 初始化时区设置，使程序能够识别容器的TZ环境变量
 * @description
 * - 从环境变量TZ获取时区设置
 * - 如果未设置TZ环境变量，使用默认时区
 * - 尝试加载指定的时区信息
 * - 如果加载失败，尝试使用固定偏移量设置时区
 * - 设置全局默认时区为识别的时区
 * @example
 * initTimeZone()
 * // 如果TZ=Asia/Shanghai，输出: 时区已设置为: Asia/Shanghai
 * // 如果未设置TZ，输出: 未设置TZ环境变量，使用默认时区
 */
func initTimeZone() {
	tz := os.Getenv("TZ")
	if tz == "" {
		fmt.Println("未设置TZ环境变量，使用默认时区")
		return
	}
	// 加载时区信息
	if location, err := time.LoadLocation(tz); err == nil {
		// 设置全局默认时区
		time.Local = location
		fmt.Printf("时区已设置为: %s\n", tz)
	} else {
		fmt.Printf("加载时区失败: %v, 尝试使用固定偏移量\n", err)
		// 备选方案：为常见时区设置固定偏移量
		setFixedOffset(tz)
	}
}

/**
 * 为常见时区设置固定偏移量
 * @param {string} tz - 时区标识符，如"Asia/Shanghai"、"America/New_York"等
 * @description
 * - 为常见时区提供固定偏移量设置
 * - 支持亚洲、欧洲、美洲等多个主要时区
 * - 使用UTC偏移量创建固定时区
 * - 如果遇到未知时区，使用默认时区并记录警告
 * - 设置全局默认时区为计算的固定偏移时区
 * @example
 * setFixedOffset("Asia/Shanghai")
 * // 输出: 已使用固定偏移量设置时区: Asia/Shanghai (UTC+8)
 */
func setFixedOffset(tz string) {
	// 创建固定偏移的时区
	var offset int
	switch tz {
	case "Asia/Shanghai", "Asia/Chongqing", "Asia/Hong_Kong", "Asia/Taipei", "PRC", "CST":
		offset = 8 * 3600 // UTC+8
	case "Asia/Tokyo", "JST":
		offset = 9 * 3600 // UTC+9
	case "Asia/Seoul", "KST":
		offset = 9 * 3600 // UTC+9
	case "Asia/Singapore":
		offset = 8 * 3600 // UTC+8
	case "Asia/Kolkata", "IST":
		offset = 5*3600 + 30*60 // UTC+5:30
	case "Europe/London", "GMT":
		offset = 0 // UTC+0
	case "Europe/Paris", "Europe/Berlin", "CET":
		offset = 1 * 3600 // UTC+1
	case "America/New_York", "EST":
		offset = -5 * 3600 // UTC-5
	case "America/Los_Angeles", "PST":
		offset = -8 * 3600 // UTC-8
	default:
		fmt.Printf("未知时区: %s, 使用默认时区\n", tz)
		return
	}

	// 创建固定偏移的时区
	time.Local = time.FixedZone(tz, offset)
	fmt.Printf("已使用固定偏移量设置时区: %s (UTC%+d)\n", tz, offset/3600)
}

/**
 * 初始化模型实例
 * @description
 * - 记录模型初始化开始日志
 * - 从配置中获取模型配置信息
 * - 调用model包的Init方法初始化所有模型
 * - 如果初始化失败，抛出panic终止程序
 * - 用于main函数中初始化模型相关组件
 * @throws
 * - 如果模型初始化失败，会导致程序panic并退出
 * @example
 * initModels()
 * // 输出日志: Initialize model instances
 */
func initModels() {
	zap.L().Info("Initialize model instances")
	if err := model.Init(config.Config.Models); err != nil {
		panic(err)
	}
}

func initStreamController() {
	zap.L().Info("Initialize the stream-controller")

	sc := stream_controller.NewStreamController()
	sc.Init()
	stream_controller.Controller = sc
}
