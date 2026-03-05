package logger

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func Test_ExampleUsage(t *testing.T) {
	// 模拟业务数据
	userID := "user_12345"
	httpMethod := "GET"
	httpURL := "/api/resource"
	latency := 15 * time.Millisecond
	status := 200
	// 模拟一个结构体，作为日志的一部分
	testStruct := struct {
		Id   string
		Name string
	}{
		Id:   "test id",
		Name: "test user",
	}
	var requestId = "01980bc0-9426-7650-8607-68d0a5b7b17d" // uuid

	// 使用结构化日志，避免字符串拼接，提高性能
	Info("HTTP request handled",
		zap.String("requestId", requestId),
		zap.String("userId", userID),
		zap.String("method", httpMethod),
		zap.String("url", httpURL),
		zap.Duration("latency", latency),
		zap.Int("status", status),
		zap.Bool("success", true),
		zap.Any("testStruct", testStruct),
	)
	Error("数据库无法链接",
		zap.String("requestId", requestId),
		zap.String("error", "数据库无法链接详情"))

	zap.L().Info("Database connected",
		zap.String("requestId", requestId),
		zap.String("error", "数据库已经连接"))
}
