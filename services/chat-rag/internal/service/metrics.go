package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"go.uber.org/zap"
)

const (
	// Base labels
	metricsBaseLabelClientID   = "client_id"
	metricsBaseLabelClientIDE  = "client_ide"
	metricsBaseLabelModel      = "model"
	metricsBaseLabelUser       = "user"
	metricsBaseLabelLoginFrom  = "login_from"
	metricsBaseLabelCaller     = "caller"
	metricsBaseLabelSender     = "sender"
	metricsBaseLabelDept1      = "dept_level1"
	metricsBaseLabelDept2      = "dept_level2"
	metricsBaseLabelDept3      = "dept_level3"
	metricsBaseLabelDept4      = "dept_level4"
	metricsBaseLabelPromptMode = "prompt_mode"

	// Label names
	metricsLabelCategory   = "category"
	metricsLabelTokenScope = "token_scope"
	metricsLabelErrorType  = "error_type"

	// Metric names
	metricRequestsTotal         = "chat_rag_requests_total"
	metricOriginalTokensTotal   = "chat_rag_original_tokens_total"
	metricCompressedTokensTotal = "chat_rag_compressed_tokens_total"
	metricFirstTokenLatency     = "chat_rag_first_token_latency_ms"
	metricWindowLatency         = "chat_rag_window_latency_ms"
	metricMainModelLatency      = "chat_rag_main_model_latency_ms"
	metricTotalLatency          = "chat_rag_total_latency_ms"
	metricResponseTokens        = "chat_rag_response_tokens_total"
	metricErrorsTotal           = "chat_rag_errors_total"
	metricTokenRatio            = "chat_rag_token_ratio"

	// Default values
	defaultCategory    = "unknown"
	defaultPromoptMode = "vibe"

	// Token scope constants
	tokenScopeSystem = "system"
	tokenScopeUser   = "user"
	tokenScopeAll    = "all"
)

// Bucket definitions
var (
	modelLatencyBuckets = []float64{
		100, 500, 1000, 2000, 5000, 10000,
		20000, 30000, 60000, 120000, 300000,
	}
)

// Base label list
var metricsBaseLabels = []string{
	metricsBaseLabelClientID,
	metricsBaseLabelClientIDE,
	metricsBaseLabelModel,
	metricsBaseLabelUser,
	metricsBaseLabelLoginFrom,
	metricsBaseLabelCaller,
	metricsBaseLabelSender,
	metricsBaseLabelDept1,
	metricsBaseLabelDept2,
	metricsBaseLabelDept3,
	metricsBaseLabelDept4,
	metricsBaseLabelPromptMode,
}

// MetricsInterface defines the interface for metrics service
type MetricsInterface interface {
	RecordChatLog(log *model.ChatLog)
	GetRegistry() *prometheus.Registry
}

// MetricsService handles Prometheus metrics collection
type MetricsService struct {
	requestsTotal         *prometheus.CounterVec
	originalTokensTotal   *prometheus.CounterVec
	compressedTokensTotal *prometheus.CounterVec
	fistTokenLatency      *prometheus.HistogramVec
	windowLatency         *prometheus.HistogramVec
	mainModelLatency      *prometheus.HistogramVec
	totalLatency          *prometheus.HistogramVec
	responseTokens        *prometheus.CounterVec
	errorsTotal           *prometheus.CounterVec
	tokenRatio            *prometheus.GaugeVec
}

// NewMetricsService creates a new metrics service
func NewMetricsService() MetricsInterface {
	ms := &MetricsService{}

	ms.requestsTotal = ms.createCounterVec(metricRequestsTotal, "Total number of chat completion requests", metricsLabelCategory)
	ms.originalTokensTotal = ms.createCounterVec(metricOriginalTokensTotal, "Total number of original tokens processed", metricsLabelTokenScope)
	ms.compressedTokensTotal = ms.createCounterVec(metricCompressedTokensTotal, "Total number of compressed tokens processed", metricsLabelTokenScope)
	ms.fistTokenLatency = ms.createHistogramVec(metricFirstTokenLatency, "Fist token received latency in milliseconds", nil, modelLatencyBuckets)
	ms.windowLatency = ms.createHistogramVec(metricWindowLatency, "Window latency in milliseconds", nil, modelLatencyBuckets)
	ms.mainModelLatency = ms.createHistogramVec(metricMainModelLatency, "Main model processing latency in milliseconds", nil, modelLatencyBuckets)
	ms.totalLatency = ms.createHistogramVec(metricTotalLatency, "Total processing latency in milliseconds", nil, modelLatencyBuckets)
	ms.responseTokens = ms.createCounterVec(metricResponseTokens, "Total number of response tokens generated")
	ms.errorsTotal = ms.createCounterVec(metricErrorsTotal, "Total number of errors encountered", metricsLabelErrorType)
	ms.tokenRatio = ms.createGaugeVec(metricTokenRatio, "Token compression ratio by scope", metricsLabelTokenScope)

	ms.registerMetrics()
	return ms
}

// createCounterVec creates a CounterVec with base labels
func (ms *MetricsService) createCounterVec(name, help string, extraLabels ...string) *prometheus.CounterVec {
	labels := metricsBaseLabels
	if len(extraLabels) > 0 {
		labels = append(labels, extraLabels...)
	}
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
}

// createHistogramVec creates a HistogramVec with base labels
func (ms *MetricsService) createHistogramVec(name, help string, extraLabels []string, buckets []float64) *prometheus.HistogramVec {
	labels := metricsBaseLabels
	if extraLabels != nil {
		labels = append(labels, extraLabels...)
	}
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)
}

// createGaugeVec creates a GaugeVec with base labels
func (ms *MetricsService) createGaugeVec(name, help string, extraLabels ...string) *prometheus.GaugeVec {
	labels := metricsBaseLabels
	if len(extraLabels) > 0 {
		labels = append(labels, extraLabels...)
	}
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
}

// registerMetrics registers all metrics
func (ms *MetricsService) registerMetrics() {
	prometheus.MustRegister(
		ms.requestsTotal,
		ms.originalTokensTotal,
		ms.compressedTokensTotal,
		ms.fistTokenLatency,
		ms.windowLatency,
		ms.mainModelLatency,
		ms.totalLatency,
		ms.responseTokens,
		ms.errorsTotal,
		ms.tokenRatio,
	)
}

// RecordChatLog records metrics from a ChatLog entry
func (ms *MetricsService) RecordChatLog(log *model.ChatLog) {
	if log == nil {
		return
	}

	labels := ms.getBaseLabels(log)
	ms.recordRequestMetrics(log, labels)
	ms.recordTokenMetrics(log, labels)
	ms.recordLatencyMetrics(log, labels)
	ms.recordResponseMetrics(log, labels)
	ms.recordErrorMetrics(log, labels)
	ms.recordTokenRatioMetrics(log, labels)
}

// recordRequestMetrics records request related metrics
func (ms *MetricsService) recordRequestMetrics(log *model.ChatLog, labels prometheus.Labels) {
	category := log.Category
	if category == "" {
		category = defaultCategory
	}
	ms.requestsTotal.With(ms.addLabel(labels, metricsLabelCategory, category)).Inc()
}

// recordTokenMetrics records token related metrics
func (ms *MetricsService) recordTokenMetrics(log *model.ChatLog, labels prometheus.Labels) {
	// Record original tokens
	ms.recordTokenCount(ms.originalTokensTotal, log.Tokens.Original, labels)

	// Record compressed tokens
	ms.recordTokenCount(ms.compressedTokensTotal, log.Tokens.Processed, labels)
}

// recordTokenCount records token count
func (ms *MetricsService) recordTokenCount(metric *prometheus.CounterVec, tokens types.TokenStats, labels prometheus.Labels) {
	record := func(scope string, count int) {
		if count < 0 {
			logger.Warn("WARNING: negative token count",
				zap.String("scope", scope),
				zap.Int("count", count),
			)
			return
		}

		if count == 0 {
			return
		}

		metric.With(ms.addLabel(labels, metricsLabelTokenScope, scope)).Add(float64(count))
	}

	record(tokenScopeSystem, tokens.SystemTokens)
	record(tokenScopeUser, tokens.UserTokens)
	record(tokenScopeAll, tokens.All)
}

// recordLatencyMetrics records latency related metrics
func (ms *MetricsService) recordLatencyMetrics(log *model.ChatLog, labels prometheus.Labels) {
	if log.Latency.MainModelLatency > 0 {
		ms.mainModelLatency.With(labels).Observe(float64(log.Latency.MainModelLatency))
	}
	if log.Latency.TotalLatency > 0 {
		ms.totalLatency.With(labels).Observe(float64(log.Latency.TotalLatency))
	}
	if log.Latency.FirstTokenLatency > 0 {
		ms.fistTokenLatency.With(labels).Observe(float64(log.Latency.FirstTokenLatency))
	}
	if log.Latency.WindowLatency > 0 {
		ms.windowLatency.With(labels).Observe(float64(log.Latency.WindowLatency))
	}
}

// recordResponseMetrics records response related metrics
func (ms *MetricsService) recordResponseMetrics(log *model.ChatLog, labels prometheus.Labels) {
	if log.Usage.CompletionTokens > 0 {
		ms.responseTokens.With(labels).Add(float64(log.Usage.CompletionTokens))
	}
}

// recordErrorMetrics records error related metrics
func (ms *MetricsService) recordErrorMetrics(log *model.ChatLog, labels prometheus.Labels) {
	for _, errorMap := range log.Error {
		for errorType, errorMessage := range errorMap {
			if errorMessage != "" {
				ms.errorsTotal.With(ms.addLabel(labels, metricsLabelErrorType, string(errorType))).Inc()
			}
		}
	}
}

// getBaseLabels creates base labels map
func (ms *MetricsService) getBaseLabels(log *model.ChatLog) prometheus.Labels {
	promptMode := string(log.Params.LlmParams.ExtraBody.PromptMode)
	if promptMode == "" {
		promptMode = defaultPromoptMode
	}

	labels := prometheus.Labels{
		metricsBaseLabelClientID:   log.Identity.ClientID,
		metricsBaseLabelClientIDE:  log.Identity.ClientIDE,
		metricsBaseLabelModel:      log.Params.Model,
		metricsBaseLabelUser:       log.Identity.UserName,
		metricsBaseLabelLoginFrom:  log.Identity.LoginFrom,
		metricsBaseLabelCaller:     log.Identity.Caller,
		metricsBaseLabelSender:     log.Identity.Sender,
		metricsBaseLabelPromptMode: promptMode,
	}

	if log.Identity.UserInfo != nil &&
		log.Identity.UserInfo.Department != nil &&
		log.Identity.UserInfo.EmployeeNumber != "" {
		labels[metricsBaseLabelDept1] = log.Identity.UserInfo.Department.Level1Dept
		labels[metricsBaseLabelDept2] = log.Identity.UserInfo.Department.Level2Dept
		labels[metricsBaseLabelDept3] = log.Identity.UserInfo.Department.Level3Dept
		labels[metricsBaseLabelDept4] = log.Identity.UserInfo.Department.Level4Dept
	} else {
		labels[metricsBaseLabelDept1] = ""
		labels[metricsBaseLabelDept2] = ""
		labels[metricsBaseLabelDept3] = ""
		labels[metricsBaseLabelDept4] = ""
	}

	return labels
}

// addLabel adds a new label to existing labels
func (ms *MetricsService) addLabel(baseLabels prometheus.Labels, key, value string) prometheus.Labels {
	// Copy original labels
	newLabels := make(prometheus.Labels, len(baseLabels)+1)
	for k, v := range baseLabels {
		newLabels[k] = v
	}
	// Add new label
	newLabels[key] = value
	return newLabels
}

// GetRegistry returns the Prometheus registry
func (ms *MetricsService) GetRegistry() *prometheus.Registry {
	return prometheus.DefaultRegisterer.(*prometheus.Registry)
}

// recordTokenRatioMetrics records token ratio related metrics
func (ms *MetricsService) recordTokenRatioMetrics(log *model.ChatLog, labels prometheus.Labels) {
	// Record system token ratio
	if log.Tokens.Ratios.SystemRatio >= 0 {
		ratioLabels := ms.addLabel(labels, metricsLabelTokenScope, tokenScopeSystem)
		ms.tokenRatio.With(ratioLabels).Set(log.Tokens.Ratios.SystemRatio)
	}

	// Record user token ratio
	if log.Tokens.Ratios.UserRatio >= 0 {
		ratioLabels := ms.addLabel(labels, metricsLabelTokenScope, tokenScopeUser)
		ms.tokenRatio.With(ratioLabels).Set(log.Tokens.Ratios.UserRatio)
	}

	// Record all token ratio
	if log.Tokens.Ratios.AllRatio >= 0 {
		ratioLabels := ms.addLabel(labels, metricsLabelTokenScope, tokenScopeAll)
		ms.tokenRatio.With(ratioLabels).Set(log.Tokens.Ratios.AllRatio)
	}
}
