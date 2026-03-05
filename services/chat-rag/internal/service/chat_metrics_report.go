package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"go.uber.org/zap"
)

type ChatMetricsReporter struct {
	ReportUrl string
	Method    string
}

func NewChatMetricsReporter(reportUrl string, method string) *ChatMetricsReporter {
	return &ChatMetricsReporter{
		ReportUrl: reportUrl,
		Method:    method,
	}
}

// RequestMetrics 表示请求指标
type RequestMetrics struct {
	SystemTokens          int `json:"system_tokens"`
	UserTokens            int `json:"user_tokens"`
	RetryNum              int `json:"retry_num"`
	ProcessedSystemTokens int `json:"processed_system_tokens"`
	ProcessedUserTokens   int `json:"processed_user_tokens"`
}

// ResponseMetrics 表示响应指标
type ResponseMetrics struct {
	Duration           float64 `json:"duration"`
	PromptTokens       int     `json:"prompt_tokens"`
	CompletionTokens   int     `json:"completion_tokens"`
	CacheTokens        int     `json:"cache_tokens"`
	FirstTokenDuration float64 `json:"first_token_duration,omitempty"`
	SlowChunk          int64   `json:"slow_chunk,omitempty"`
	ChunkPerSecond     float64 `json:"chunk_per_second,omitempty"`
	ErrorCode          string  `json:"error_code,omitempty"`
}

// Label 表示标签信息
type Label struct {
	ClientVersion      string `json:"client_version,omitempty"`
	RequestTime        string `json:"request_time,omitempty"`
	ForwardRequestTime string `json:"forward_request_time,omitempty"`
	EndTime            string `json:"end_time,omitempty"`
	Mode               string `json:"mode,omitempty"`
	Model              string `json:"model,omitempty"`
}

// MetricsReport 表示完整的指标上报数据
type MetricsReport struct {
	RequestID       string          `json:"request_id"`
	RequestMetrics  RequestMetrics  `json:"request_metrics"`
	ResponseMetrics ResponseMetrics `json:"response_metrics"`
	Label           Label           `json:"label"`
}

// ReportMetrics 上报聊天指标,errors 为了防止并发问题,单独处理
func (mr *ChatMetricsReporter) ReportMetrics(chatLog *model.ChatLog, errors ...string) {
	if mr.ReportUrl == "" {
		logger.Debug("metrics report url is empty, skip reporting")
		return
	}

	report := mr.convertChatLogToReport(chatLog, errors...)

	if err := mr.sendReport(report, chatLog.Identity.AuthToken); err != nil {
		logger.Error("failed to report metrics", zap.String("request_id", chatLog.Identity.RequestID), zap.Error(err))
	}
}

// convertChatLogToReport 将 ChatLog 转换为 MetricsReport
func (mr *ChatMetricsReporter) convertChatLogToReport(chatLog *model.ChatLog, errors ...string) *MetricsReport {
	report := &MetricsReport{
		RequestID:       chatLog.Identity.RequestID,
		RequestMetrics:  mr.buildRequestMetrics(chatLog),
		ResponseMetrics: mr.buildResponseMetrics(chatLog, errors),
		Label:           mr.buildLabel(chatLog),
	}

	return report
}

// buildRequestMetrics 构建请求指标
func (mr *ChatMetricsReporter) buildRequestMetrics(chatLog *model.ChatLog) RequestMetrics {
	// 处理后系统提示词长度
	processedSystemTokens := 0
	if chatLog.IsPromptProceed {
		processedSystemTokens = chatLog.Tokens.Processed.SystemTokens
	}

	// 处理后用户提示词长度
	processedUserTokens := 0
	if chatLog.IsPromptProceed {
		processedUserTokens = chatLog.Tokens.Processed.UserTokens
	}

	// 重试次数
	retryNum := 0 // 当前版本忽略

	return RequestMetrics{
		SystemTokens:          chatLog.Tokens.Original.SystemTokens,
		UserTokens:            chatLog.Tokens.Original.UserTokens,
		RetryNum:              retryNum,
		ProcessedSystemTokens: processedSystemTokens,
		ProcessedUserTokens:   processedUserTokens,
	}
}

// buildResponseMetrics 构建响应指标
func (mr *ChatMetricsReporter) buildResponseMetrics(chatLog *model.ChatLog, errors []string) ResponseMetrics {
	metrics := ResponseMetrics{
		Duration:         float64(chatLog.Latency.TotalLatency),
		PromptTokens:     chatLog.Usage.PromptTokens,
		CompletionTokens: chatLog.Usage.CompletionTokens,
		CacheTokens:      0, // 默认为0，如果后续有缓存tokens数据可以添加
	}

	// 首token时长 (ms)
	if chatLog.Latency.FirstTokenLatency > 0 {
		metrics.FirstTokenDuration = float64(chatLog.Latency.FirstTokenLatency)
	}

	// 错误类型
	if len(errors) > 0 {
		metrics.ErrorCode = errors[0] // 只取第一个错误类型
	}

	// // 计算 chunk_per_second (每秒处理的chunk数)
	// if chatLog.Latency.TotalLatency > 0 && chatLog.Usage.CompletionTokens > 0 {
	// 	// 转换为秒
	// 	durationInSeconds := float64(chatLog.Latency.TotalLatency) / 1000.0
	// 	metrics.ChunkPerSecond = float64(chatLog.Usage.CompletionTokens) / durationInSeconds
	// }

	// 最耗时的chunk (可以根据实际需求启用)
	// slowestChunk := int64(0)
	// for _, tool := range chatLog.ToolCalls {
	// 	if tool.Latency > slowestChunk {
	// 		slowestChunk = tool.Latency
	// 	}
	// }
	// if slowestChunk > 0 {
	// 	metrics.SlowChunk = slowestChunk
	// }

	return metrics
}

// buildLabel 构建标签
func (mr *ChatMetricsReporter) buildLabel(chatLog *model.ChatLog) Label {
	label := Label{
		ClientVersion: chatLog.Identity.ClientVersion,
		Model:         chatLog.Params.Model,
	}

	// 请求时间 - 使用chatLog的时间戳
	if !chatLog.Timestamp.IsZero() {
		label.RequestTime = chatLog.Timestamp.Format(time.RFC3339)
	}

	// 转发时间 - 如果有首token延迟，可以计算转发时间
	if chatLog.Latency.FirstTokenLatency > 0 {
		forwardTime := chatLog.Timestamp.Add(time.Duration(chatLog.Latency.FirstTokenLatency) * time.Millisecond)
		label.ForwardRequestTime = forwardTime.Format(time.RFC3339)
	}

	// 结束时间 - 使用总延迟计算
	if chatLog.Latency.TotalLatency > 0 {
		endTime := chatLog.Timestamp.Add(time.Duration(chatLog.Latency.TotalLatency) * time.Millisecond)
		label.EndTime = endTime.Format(time.RFC3339)
	}

	// 模式 - 从请求参数中提取
	if chatLog.Params.LlmParams.ExtraBody.Mode != "" {
		label.Mode = chatLog.Params.LlmParams.ExtraBody.Mode
	}

	return label
}

// sendReport 发送指标报告
func (mr *ChatMetricsReporter) sendReport(report *MetricsReport, authToken string) error {
	jsonData, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	req, err := http.NewRequest(mr.Method, mr.ReportUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// 添加认证头
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	client := &http.Client{
		Timeout: 10 * time.Second, // 设置请求超时时间,防止大量阻塞
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	logger.Debug("metrics reported successfully",
		zap.String("request_id", report.RequestID),
	)

	return nil
}

