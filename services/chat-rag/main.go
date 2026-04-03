package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zgsm-ai/chat-rag/internal/api"
	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"go.uber.org/zap"
)

// main is the entry point of the chat-rag service
func main() {
	// Load config
	var configFile string
	flag.StringVar(&configFile, "f", "etc/chat-api.yaml", "the config file")
	flag.Parse()

	c := config.MustLoadConfig(configFile)

	// Create gin engine
	router := gin.Default()

	// Initialize service context
	ctx := bootstrap.NewServiceContext(c)

	// Register routes
	api.RegisterHandlers(router, ctx)

	// Create HTTP server with graceful shutdown support
	server := &http.Server{
		Addr:    c.Host + ":" + strconv.Itoa(c.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("server starting",
			zap.String("address", server.Addr),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server",
				zap.Error(err),
			)
			panic("Failed to start server: " + err.Error())
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// Stop service context (gracefully shutdown all services)
	logger.Info("Stopping service context...")
	ctx.Stop()

	// Shutdown HTTP server
	logger.Info("Shutting down HTTP server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown",
			zap.Error(err),
		)
	}

	logger.Info("Server exited")
}
