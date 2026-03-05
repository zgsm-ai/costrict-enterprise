package server

import (
	"code-completion/pkg/logger"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

/**
 * 服务器结构体
 * @description
 * - 封装HTTP服务器实例
 * - 包含日志记录器引用
 * - 提供服务器启动、停止等管理功能
 * - 支持优雅关闭机制
 * @example
 * srv := NewServer(":8080", router)
 * err := srv.Start()
 */
type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

/**
 * 创建新的服务器实例
 * @param {string} addr - 服务器监听地址，格式为"host:port"
 * @param {*gin.Engine} router - Gin路由引擎实例，用于处理HTTP请求
 * @returns {*Server} 返回配置好的服务器实例指针
 * @description
 * - 创建并初始化HTTP服务器配置
 * - 设置服务器监听地址和请求处理器
 * - 初始化日志记录器引用
 * - 返回可用于启动服务器的实例
 * @example
 * router := gin.Default()
 * srv := NewServer("127.0.0.1:8080", router)
 */
func NewServer(addr string, router *gin.Engine) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: router,
		},
		logger: logger.Logger,
	}
}

/**
 * 启动服务器
 * @returns {error} 返回服务器启动或关闭过程中的错误，成功返回nil
 * @description
 * - 在goroutine中启动HTTP服务器
 * - 监听系统中断信号(SIGINT, SIGTERM)
 * - 收到中断信号后执行优雅关闭
 * - 设置30秒超时用于完成正在处理的请求
 * - 记录服务器启动和关闭的日志信息
 * @throws
 * - HTTP服务器启动失败时记录fatal日志
 * - 服务器关闭失败时返回错误
 * @example
 * srv := NewServer(":8080", router)
 * if err := srv.Start(); err != nil {
 *     log.Fatal("服务器启动失败:", err)
 * }
 */
func (s *Server) Start() error {
	s.logger.Info("启动Gin服务器", zap.String("addr", s.httpServer.Addr))

	// 启动HTTP服务器
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("正在关闭服务器...")

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("服务器关闭失败", zap.Error(err))
		return err
	}

	s.logger.Info("服务器已优雅关闭")
	return nil
}

/*
*
* 停止服务器
* @returns {error} 返回服务器关闭过程中的错误，成功返回nil
* @description
* - 创建5秒超时的上下文
* - 调用HTTP服务器的Shutdown方法进行优雅关闭
* - 等待正在处理的请求完成或超时
* - 用于主动停止服务器运行
* @example
* srv := NewServer(":8080", router)
// 启动服务器...

	if err := srv.Stop(); err != nil {
	    log.Println("服务器停止失败:", err)
	}
*/
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
