package logic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/client"
	"github.com/zgsm-ai/chat-rag/internal/functions"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/promptflow"
	"github.com/zgsm-ai/chat-rag/internal/promptflow/ds"
	"github.com/zgsm-ai/chat-rag/internal/router"
	"github.com/zgsm-ai/chat-rag/internal/timeout"
	"github.com/zgsm-ai/chat-rag/internal/tokenizer"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"github.com/zgsm-ai/chat-rag/internal/utils"
)

type ChatCompletionLogic struct {
	ctx             context.Context
	svcCtx          *bootstrap.ServiceContext
	request         *types.ChatCompletionRequest
	writer          http.ResponseWriter
	headers         *http.Header
	identity        *model.Identity
	responseHandler *ResponseHandler
	toolExecutor    functions.ToolExecutor
	usage           *types.Usage
	orderedModels   []string
	streamCommitted bool
	originalModel   string
}

func NewChatCompletionLogic(
	ctx context.Context,
	svcCtx *bootstrap.ServiceContext,
	request *types.ChatCompletionRequest,
	writer http.ResponseWriter,
	headers *http.Header,
	identity *model.Identity,
) *ChatCompletionLogic {
	return &ChatCompletionLogic{
		ctx:             ctx,
		svcCtx:          svcCtx,
		identity:        identity,
		responseHandler: NewResponseHandler(ctx, svcCtx),
		request:         request,
		writer:          writer,
		headers:         headers,
		toolExecutor:    svcCtx.ToolExecutor,
		originalModel:   request.Model,
	}
}

const (
	MaxToolCallDepth    = 6
	MaxToolResultLength = 100_000
)

// processRequest handles common request processing logic
func (l *ChatCompletionLogic) processRequest() (*model.ChatLog, *ds.ProcessedPrompt, error) {
	logger.InfoC(l.ctx, "starting to process request",
		zap.String("user", l.identity.UserName), zap.String("model", l.request.Model))
	startTime := time.Now()

	// Set request priority for valid VIP users when feature is enabled (VIP > 0 and not expired)
	if l.svcCtx.Config.VIPPriority.Enabled && l.identity.UserInfo != nil && l.identity.UserInfo.Vip > 0 {
		isVipValid := l.identity.UserInfo.VipExpire == nil || time.Now().Before(*l.identity.UserInfo.VipExpire)

		if isVipValid {
			priority := 10
			l.request.Priority = &priority
			logger.InfoC(l.ctx, "vip user detected, set priority",
				zap.String("user", l.identity.UserName),
				zap.Int("priority", priority),
				zap.Int("vip_level", l.identity.UserInfo.Vip))
		}
	}

	// Initialize chat log
	chatLog := l.newChatLog(startTime)

	promptArranger := promptflow.NewPromptProcessor(
		l.ctx,
		l.svcCtx,
		l.request.ExtraBody.PromptMode,
		l.headers,
		l.identity,
		l.request.Model,
	)
	processedPrompt, err := promptArranger.Arrange(l.request.Messages)
	if err != nil {
		return chatLog, nil, fmt.Errorf("failed to process prompt:\n %w", err)
	}

	// Update chat log with processed prompt info
	l.updateChatLog(chatLog, processedPrompt)
	return chatLog, processedPrompt, nil
}

func (l *ChatCompletionLogic) newChatLog(startTime time.Time) *model.ChatLog {
	userTokens := l.countTokensInMessages(utils.GetUserMsgs(l.request.Messages))
	allTokens := l.countTokensInMessages(l.request.Messages)

	// Create a deep copy of the original messages to avoid reference issues
	// originalPrompt := make([]types.Message, len(l.request.Messages))
	// copy(originalPrompt, l.request.Messages)

	modelName := l.originalModel
	if modelName == "" {
		modelName = l.request.Model
	}

	return &model.ChatLog{
		Identity:  *l.identity,
		Timestamp: startTime,
		Params: model.RequestParams{
			Model:     modelName,
			LlmParams: l.request.LLMRequestParams,
		},
		Tokens: types.TokenMetrics{
			Original: types.TokenStats{
				SystemTokens: allTokens - userTokens,
				UserTokens:   userTokens,
				All:          allTokens,
			},
		},
		// OriginalPrompt: originalPrompt,
	}
}

// updateChatLog updates the chat log with information from the processed prompt
func (l *ChatCompletionLogic) updateChatLog(chatLog *model.ChatLog, processedPrompt *ds.ProcessedPrompt) {
	// Update log with processed prompt info
	allTokens := l.countTokensInMessages(processedPrompt.Messages)
	userTokens := l.countTokensInMessages(utils.GetUserMsgs(processedPrompt.Messages))

	chatLog.Tokens.Processed = types.TokenStats{
		SystemTokens: allTokens - userTokens,
		UserTokens:   userTokens,
		All:          allTokens,
	}
	// Calculate ratios after setting processed tokens
	chatLog.Tokens.Ratios = processedPrompt.TokenMetrics.Ratios

	chatLog.ProcessedPrompt = processedPrompt.Messages
	chatLog.Agent = processedPrompt.Agent
}

func (l *ChatCompletionLogic) logCompletion(chatLog *model.ChatLog) {
	chatLog.Latency.TotalLatency = time.Since(chatLog.Timestamp).Milliseconds()
	if l.svcCtx.LoggerService != nil {
		l.svcCtx.LoggerService.LogAsync(chatLog, l.headers)
	}
}

// ChatCompletion handles chat completion requests
func (l *ChatCompletionLogic) ChatCompletion() (resp *types.ChatCompletionResponse, err error) {
	// Router: select model before prompt processing & LLM client creation
	origModel := l.request.Model
	if l.svcCtx.Config.Router != nil && l.svcCtx.Config.Router.Enabled && strings.EqualFold(l.request.Model, "auto") {
		logger.InfoC(l.ctx, "semantic router: auto mode routing start",
			zap.String("strategy", l.svcCtx.Config.Router.Strategy),
		)
		// Use cached strategy instance to maintain state across requests (e.g., round-robin weights)
		if runner := l.getOrCreateRouterStrategy(); runner != nil {
			selected, current, ordered, rerr := runner.Run(l.ctx, l.svcCtx, l.headers, l.request)
			if rerr == nil && selected != "" {
				l.request.Model = selected
				l.orderedModels = ordered
				// mark original model via request header for upstream
				if l.headers != nil && strings.EqualFold(origModel, "auto") {
					l.headers.Set(types.HeaderOriginalModel, "Auto")
				}
				if l.writer != nil {
					l.writer.Header().Set(types.HeaderSelectLLm, selected)
					if current != "" {
						safe := sanitizeHeaderValue(current)
						if safe != "" {
							encodedCur := base64.StdEncoding.EncodeToString([]byte(safe))
							if encodedCur != "" {
								l.writer.Header().Set(types.HeaderUserInput, encodedCur)
							}
						}
					}
				}
				logger.InfoC(l.ctx, "semantic router: auto mode routing selected",
					zap.String("selected_model", selected),
					zap.Int("user_input_len", len([]byte(current))),
				)
			}
		}
	}

	chatLog, processedPrompt, err := l.processRequest()

	defer l.logCompletion(chatLog)

	if err == nil {
		l.request.Messages = processedPrompt.Messages
		chatLog.IsPromptProceed = true
	} else {
		err := fmt.Errorf("ChatCompletion failed to process request:\n%w", err)
		logger.ErrorC(l.ctx, "failed to process request", zap.Error(err))
		chatLog.AddError(types.ErrServerError, err)
		chatLog.IsPromptProceed = false
	}

	// Create shared idle tracker for the entire request (both retry and degradation)
	_, _, _, totalIdleTimeout := l.getRetryConfig()
	idleTracker := timeout.NewIdleTracker(totalIdleTimeout)

	modelStart := time.Now()
	var response types.ChatCompletionResponse
	// Smart degradation when ordered models are available
	if len(l.orderedModels) > 0 {
		logger.InfoC(l.ctx, "degradation: attempting ordered models",
			zap.Strings("ordered", l.orderedModels),
		)
		resp, derr := l.callWithDegradation(l.request.LLMRequestParams, idleTracker)
		if derr != nil {
			chatLog.AddError(types.ErrApiError, derr)
			return nil, derr
		}
		response = resp
	} else {
		// Fallback to single model with retry
		var err2 error
		response, err2 = l.callModelWithRetry(l.request.Model, l.request.LLMRequestParams, idleTracker)
		if err2 != nil {
			if l.isContextLengthError(err2) {
				logger.ErrorC(l.ctx, "Input context too long, exceeded limit.", zap.Error(err2))
				lengthErr := types.NewContextTooLongError()
				l.responseHandler.sendSSEError(l.ctx, l.writer, lengthErr)
				chatLog.AddError(types.ErrContextExceeded, lengthErr)
				return nil, lengthErr
			}
			chatLog.AddError(types.ErrApiError, err2)
			return nil, err2
		}
	}

	chatLog.Latency.MainModelLatency = time.Since(modelStart).Milliseconds()

	// Extract response content and usage information
	l.responseHandler.extractResponseInfo(chatLog, &response)
	return &response, nil
}

// getRetryConfig returns retry and timeout configuration based on the current mode
func (l *ChatCompletionLogic) getRetryConfig() (maxRetryCount int, retryInterval time.Duration, idleTimeout time.Duration, totalIdleTimeout time.Duration) {
	isAutoMode := len(l.orderedModels) > 0
	if isAutoMode {
		// Model degradation mode: use routing configuration based on strategy
		if l.svcCtx.Config.Router != nil && l.svcCtx.Config.Router.Strategy == "priority" {
			// Priority strategy: use priority configuration
			maxRetryCount = l.svcCtx.Config.Router.Priority.MaxRetryCount
			retryInterval = time.Duration(l.svcCtx.Config.Router.Priority.RetryIntervalMs) * time.Millisecond
			idleTimeout = time.Duration(l.svcCtx.Config.Router.Priority.IdleTimeoutMs) * time.Millisecond
			totalIdleTimeout = time.Duration(l.svcCtx.Config.Router.Priority.TotalIdleTimeoutMs) * time.Millisecond
		} else {
			// Semantic strategy: use semantic routing configuration
			maxRetryCount = l.svcCtx.Config.Router.Semantic.Routing.MaxRetryCount
			retryInterval = time.Duration(l.svcCtx.Config.Router.Semantic.Routing.RetryIntervalMs) * time.Millisecond
			idleTimeout = time.Duration(l.svcCtx.Config.Router.Semantic.Routing.IdleTimeoutMs) * time.Millisecond
			totalIdleTimeout = time.Duration(l.svcCtx.Config.Router.Semantic.Routing.TotalIdleTimeoutMs) * time.Millisecond
		}
	} else {
		// Regular mode: use llmTimeout configuration
		maxRetryCount = l.svcCtx.Config.LLMTimeout.MaxRetryCount
		retryInterval = time.Duration(l.svcCtx.Config.LLMTimeout.RetryIntervalMs) * time.Millisecond
		idleTimeout = time.Duration(l.svcCtx.Config.LLMTimeout.IdleTimeoutMs) * time.Millisecond
		totalIdleTimeout = time.Duration(l.svcCtx.Config.LLMTimeout.TotalIdleTimeoutMs) * time.Millisecond
	}
	return
}

// ChatCompletionStream handles streaming chat completion with SSE
func (l *ChatCompletionLogic) ChatCompletionStream() error {
	// Router: select model before streaming LLM client creation
	origModel := l.request.Model
	if l.svcCtx.Config.Router != nil && l.svcCtx.Config.Router.Enabled && strings.EqualFold(l.request.Model, "auto") {
		logger.InfoC(l.ctx, "semantic router: auto mode routing start",
			zap.String("strategy", l.svcCtx.Config.Router.Strategy),
		)
		// Use cached strategy instance to maintain state across requests (e.g., round-robin weights)
		if runner := l.getOrCreateRouterStrategy(); runner != nil {
			selected, current, ordered, rerr := runner.Run(l.ctx, l.svcCtx, l.headers, l.request)
			if rerr == nil && selected != "" {
				l.request.Model = selected
				l.orderedModels = ordered
				// mark original model via request header for upstream
				if l.headers != nil && strings.EqualFold(origModel, "auto") {
					l.headers.Set(types.HeaderOriginalModel, "Auto")
				}
				if l.writer != nil {
					l.writer.Header().Set(types.HeaderSelectLLm, selected)
					if current != "" {
						safe := sanitizeHeaderValue(current)
						if safe != "" {
							encodedCur := base64.StdEncoding.EncodeToString([]byte(safe))
							if encodedCur != "" {
								l.writer.Header().Set(types.HeaderUserInput, encodedCur)
							}
						}
					}
				}
				logger.InfoC(l.ctx, "semantic router: auto mode routing selected",
					zap.String("selected_model", selected),
					zap.Int("user_input_len", len([]byte(current))),
				)
			}
		}
	}

	chatLog, processedPrompt, err := l.processRequest()

	defer l.logCompletion(chatLog)

	if err == nil {
		l.request.Messages = processedPrompt.Messages
		chatLog.IsPromptProceed = true
	} else {
		err := fmt.Errorf("ChatCompletionStream failed to process request: %w", err)
		logger.ErrorC(l.ctx, "failed to process request in streaming", zap.Error(err))
		chatLog.AddError(types.ErrServerError, err)
		chatLog.IsPromptProceed = false
	}

	flusher, ok := l.writer.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	// Create shared idle tracker for the entire request (both retry and degradation)
	_, _, _, totalIdleTimeout := l.getRetryConfig()
	idleTracker := timeout.NewIdleTracker(totalIdleTimeout)

	// Streaming degradation only for auto mode (when orderedModels is present)
	if len(l.orderedModels) == 0 {
		// No degradation list → single model streaming (non-auto path) with retry
		maxRetryCount, retryInterval, _, _ := l.getRetryConfig()
		var lastErr error
		for attempt := 0; attempt <= maxRetryCount; attempt++ {
			logger.InfoC(l.ctx, "single-model retry(stream): attempting model",
				zap.String("model", l.request.Model),
				zap.Int("attempt", attempt+1),
				zap.Int("maxRetries", maxRetryCount),
			)

			llmClient, err := client.NewLLMClient(l.svcCtx.Config.LLM, l.svcCtx.Config.LLMTimeout, l.request.Model, l.headers)
			if err != nil {
				lastErr = err
				l.responseHandler.sendSSEError(l.ctx, l.writer, err)
				chatLog.AddError(types.ErrServerError, err)
				return fmt.Errorf("LLM client creation failed: %w", err)
			}
			llmClient.SetTools(processedPrompt.Tools)
			l.streamCommitted = false

			err = l.handleStreamingWithTools(l.ctx, llmClient, flusher, chatLog, MaxToolCallDepth, idleTracker)
			if err == nil {
				return nil
			}

			lastErr = err
			if l.streamCommitted {
				return l.handleStreamError(err, chatLog)
			}

			retryable := isRetryableAPIError(err)
			logger.WarnC(l.ctx, "single-model retry(stream): attempt failed before first token",
				zap.String("model", l.request.Model),
				zap.Bool("retryable", retryable),
				zap.Error(err),
			)
			if retryable && attempt < maxRetryCount {
				// Check if we have enough idle budget remaining for retry
				remainingIdleBudget := idleTracker.Remaining()
				minRequiredBudget := retryInterval
				if remainingIdleBudget < minRequiredBudget {
					logger.WarnC(l.ctx, "single-model retry(stream): insufficient idle budget for retry",
						zap.Duration("remainingIdleBudget", remainingIdleBudget),
						zap.Duration("minRequiredBudget", minRequiredBudget))
					break
				}
				time.Sleep(retryInterval)
				continue
			}

			break
		}
		return l.handleStreamError(lastErr, chatLog)
	}

	// Degradation enabled (auto mode): only switch model if failure occurs before first token
	models := l.orderedModels

	maxRetryCount, retryInterval, _, _ := l.getRetryConfig()
	var lastErr error
	for _, modelName := range models {
		// Update header immediately when switching to a different model in auto mode
		if l.writer != nil {
			l.writer.Header().Set(types.HeaderSelectLLm, modelName)
		}

		attempt := 0
		for attempt <= maxRetryCount {
			logger.InfoC(l.ctx, "degradation(stream): attempting model",
				zap.String("model", modelName),
				zap.Int("attempt", attempt+1),
				zap.Int("maxRetries", maxRetryCount),
			)

			l.request.Model = modelName
			l.streamCommitted = false
			llmClient, err := client.NewLLMClient(l.svcCtx.Config.LLM, l.svcCtx.Config.LLMTimeout, modelName, l.headers)
			if err != nil {
				lastErr = err
				logger.WarnC(l.ctx, "degradation(stream): failed to create llm client",
					zap.String("model", modelName), zap.Error(err))
				break
			}
			llmClient.SetTools(processedPrompt.Tools)

			err = l.handleStreamingWithTools(l.ctx, llmClient, flusher, chatLog, MaxToolCallDepth, idleTracker)
			if err == nil {
				return nil
			}

			lastErr = err
			if l.streamCommitted {
				// Already started streaming; report error to client and stop
				return l.handleStreamError(err, chatLog)
			}

			retryable := isRetryableAPIError(err)
			logger.WarnC(l.ctx, "degradation(stream): attempt failed before first token",
				zap.String("model", modelName),
				zap.Bool("retryable", retryable),
				zap.Error(err),
			)
			if retryable && attempt < maxRetryCount {
				// Check if we have enough idle budget remaining for retry
				remainingIdleBudget := idleTracker.Remaining()
				minRequiredBudget := retryInterval
				if remainingIdleBudget < minRequiredBudget {
					logger.WarnC(l.ctx, "degradation(stream): insufficient idle budget for retry",
						zap.String("model", modelName),
						zap.Duration("remainingIdleBudget", remainingIdleBudget),
						zap.Duration("minRequiredBudget", minRequiredBudget))
					break
				}
				time.Sleep(retryInterval)
				attempt++
				continue
			}
			break
		}

		// If the error is context canceled, stop trying other models
		if errors.Is(lastErr, context.Canceled) || errors.Is(l.ctx.Err(), context.Canceled) {
			break
		}
	}
	return l.handleStreamError(lastErr, chatLog)
}

// streamState holds the state for streaming processing
type streamState struct {
	window       []string // Window of streamed content used for detect tools
	windowSize   int
	toolDetected bool
	toolName     string
	fullContent  strings.Builder
	response     *types.ChatCompletionResponse
	modelStart   time.Time
	firstToken   bool // Flag to track if first token has been received
	windowSent   bool // Flag to track if first token has been sent to client
}

func newStreamState() *streamState {
	return &streamState{
		windowSize: 6,
		modelStart: time.Now(),
		firstToken: true, // Initialize as true to detect first token
	}
}

func (l *ChatCompletionLogic) handleStreamingWithTools(
	ctx context.Context,
	llmClient client.LLMInterface,
	flusher http.Flusher,
	chatLog *model.ChatLog,
	remainingDepth int,
	idleTracker *timeout.IdleTracker,
) error {
	logger.InfoC(ctx, "starting to handle streaming with tools",
		zap.Int("remainingDepth", remainingDepth),
		zap.Int("MaxToolCallDepth", MaxToolCallDepth),
		zap.String("promptMode", string(l.request.ExtraBody.PromptMode)),
	)

	// If raw mode, directly pass through results to client
	if l.request.ExtraBody.PromptMode == types.Raw {
		return l.handleRawModeStream(ctx, llmClient, flusher, chatLog, idleTracker)
	}

	// If Tools or Functions are provided, also use raw mode for direct tool handling
	hasTools := false
	if tools, ok := l.request.Extra["tools"]; ok {
		if toolsSlice, ok := tools.([]any); ok && len(toolsSlice) > 0 {
			hasTools = true
		}
	}
	hasFunctions := false
	if functions, ok := l.request.Extra["functions"]; ok {
		if functionsSlice, ok := functions.([]any); ok && len(functionsSlice) > 0 {
			hasFunctions = true
		}
	}

	if hasTools || hasFunctions {
		logger.InfoC(ctx, "received function call in streaming request")
		return l.handleRawModeStream(ctx, llmClient, flusher, chatLog, idleTracker)
	}

	state := newStreamState()

	// Phase 1: Process streaming response
	toolDetected, err := l.processStream(ctx, llmClient, flusher, state, remainingDepth, chatLog, idleTracker)
	if err != nil {
		// Do not send SSE error here; let caller decide based on commit status
		return err
	}

	// Phase 2: Handle tool execution or complete response
	if toolDetected {
		return l.handleToolExecution(ctx, llmClient, flusher, chatLog, state, remainingDepth, idleTracker)
	}

	return l.completeStreamResponse(flusher, chatLog, state)
}

// processStream handles the streaming response processing
func (l *ChatCompletionLogic) processStream(
	ctx context.Context,
	llmClient client.LLMInterface,
	flusher http.Flusher,
	state *streamState,
	remainingDepth int,
	chatLog *model.ChatLog,
	idleTracker *timeout.IdleTracker,
) (bool, error) {
	// Use the provided shared idle tracker instead of creating a new one
	_, _, idleTimeout, _ := l.getRetryConfig()
	timerCtx, cancel, idleTimer := timeout.NewIdleTimer(ctx, idleTimeout, idleTracker)
	defer func() {
		idleTimer.Stop()
		cancel()
	}()

	err := llmClient.ChatLLMWithMessagesStreamRaw(timerCtx, l.request.LLMRequestParams, idleTimer, func(llmResp client.LLMResponse) error {
		l.handleResonseHeaders(llmResp.Header, types.ResponseHeadersToForward, chatLog)

		return l.handleStreamChunk(ctx, flusher, llmResp.ResonseLine, state, remainingDepth, chatLog, idleTimer)
	})
	if c, ok := llmClient.(*client.LLMClient); ok {
		streamState := c.StreamChunkInfo
		if streamState != nil {
			chatLog.Latency.ChunkInfo = &model.StreamChunkInfo{
				ChunkTotal:   streamState.Count,
				IntervalAvg:  float32(streamState.Mean),
				IntervalMax:  float32(streamState.Max),
				IntervalMin:  float32(streamState.Min),
				P50:          float32(streamState.P50),
				P95:          float32(streamState.P95),
				P99:          float32(streamState.P99),
				StdDeviation: float32(streamState.StdDev),
				Variance:     float32(streamState.Variance),
			}
		}
	}
	return state.toolDetected, err
}

// handleResonseHeaders Set the specified request header to the response
func (l *ChatCompletionLogic) handleResonseHeaders(header *http.Header, requiredHeaders []string, chatLog *model.ChatLog) {
	for _, headerName := range requiredHeaders {
		if headerValue := header.Get(headerName); headerValue != "" {
			if l.writer.Header().Get(headerName) != "" {
				continue
			}

			l.writer.Header().Set(headerName, headerValue)
			chatLog.ResponseHeaders = append(
				chatLog.ResponseHeaders,
				map[string]string{headerName: headerValue},
			)
			logger.InfoC(l.ctx, "Response header setted",
				zap.String("header", headerName), zap.String("value", headerValue))
		}
	}
}

// handleStreamChunk processes individual streaming chunks
func (l *ChatCompletionLogic) handleStreamChunk(
	ctx context.Context,
	flusher http.Flusher,
	rawLine string,
	state *streamState,
	remainingDepth int,
	chatLog *model.ChatLog,
	idleTimer *timeout.IdleTimer,
) error {
	content, usage, resp := l.responseHandler.extractStreamingData(rawLine)
	if resp != nil {
		state.response = resp
	}
	if usage != nil {
		l.usage = usage
	}
	if content == "" {
		return l.sendRawLine(flusher, rawLine)
	}

	// Log first token response
	if state.firstToken && content != "[DONE]" {
		// Mark streaming committed and set selected model header
		l.streamCommitted = true
		if l.writer != nil && len(l.orderedModels) > 0 {
			l.writer.Header().Set(types.HeaderSelectLLm, l.request.Model)
		}
		firstTokenLatency := time.Since(state.modelStart)
		chatLog.Latency.FirstTokenLatency = firstTokenLatency.Milliseconds()
		logger.InfoC(ctx, "[first-token] first token received, and response",
			zap.String("model", l.request.Model), zap.Duration("firstTokenLatency", firstTokenLatency))
		state.firstToken = false

		// 通知 idleTimer 已接收首token（新增）
		idleTimer.SetFirstTokenReceived()

		if err := l.sendStreamContent(flusher, state.response, "\n"); err != nil {
			return err
		}
	}

	// Add to window and complete content
	state.window = append(state.window, content)
	if content != "[DONE]" {
		state.fullContent.WriteString(content)
	}

	// Check for tool detection
	if !state.toolDetected && l.toolExecutor != nil && remainingDepth > 0 &&
		l.svcCtx.Config.Tools != nil && !l.svcCtx.Config.Tools.DisableTools {
		if err := l.detectAndHandleTool(ctx, flusher, state); err != nil {
			return err
		}
	}

	// Send content beyond window
	if !state.toolDetected && len(state.window) >= state.windowSize {
		// Log window tokens token sent to client
		if !state.windowSent {
			state.windowSent = true
			windowLatency := time.Since(state.modelStart)
			chatLog.Latency.WindowLatency = windowLatency.Milliseconds()
			logger.InfoC(ctx, "first window tokens sent to client",
				zap.Duration("firstWindowTokenLatency", windowLatency))
		}

		if err := l.sendStreamContent(flusher, state.response, state.window[0]); err != nil {
			return err
		}
		state.window = state.window[1:]
	}

	return nil
}

// detectAndHandleTool handles tool detection and pre-tool content sending
func (l *ChatCompletionLogic) detectAndHandleTool(ctx context.Context, flusher http.Flusher, state *streamState) error {
	currentContent := strings.Join(state.window, "")
	hasTool, name := l.toolExecutor.DetectTools(ctx, currentContent)

	if !hasTool {
		return nil
	}

	state.toolDetected = true
	state.toolName = name
	logger.InfoC(ctx, "detected server xml tool", zap.String("name", name))

	// Send content before tool call
	toolStartIndex := strings.Index(currentContent, "<"+name+">")
	if toolStartIndex > 0 {
		preToolContent := currentContent[:toolStartIndex]
		if err := l.sendStreamContent(flusher, state.response, preToolContent); err != nil {
			logger.ErrorC(ctx, "failed to sendStreamContent when detecting tool",
				zap.String("preToolContent", preToolContent), zap.Error(err))
			return err
		}
	}

	state.window = []string{currentContent[toolStartIndex:]}
	return nil
}

// handleToolExecution executes the detected tool and continues processing
func (l *ChatCompletionLogic) handleToolExecution(
	ctx context.Context,
	llmClient client.LLMInterface,
	flusher http.Flusher,
	chatLog *model.ChatLog,
	state *streamState,
	remainingDepth int,
	idleTracker *timeout.IdleTracker,
) error {
	logger.InfoC(ctx, "starting to call tool", zap.String("name", state.toolName))
	toolContent := strings.Join(state.window, "")
	toolCall := model.ToolCall{
		ToolName:  state.toolName,
		ToolInput: toolContent,
	}

	l.updateToolStatus(state.toolName, types.ToolStatusRunning)
	// Send tool use information to client page
	if err := l.sendStreamContent(flusher, state.response,
		fmt.Sprintf("%s`%s` %s", types.StrFilterToolSearchStart, state.toolName,
			types.StrFilterToolSearchEnd)); err != nil {
		return err
	}

	// wait client to refesh content
	for i := 0; i < 5; i++ {
		if err := l.sendStreamContent(flusher, state.response, "."); err != nil {
			return err
		}
		time.Sleep(600 * time.Millisecond)
	}

	// execute and record tool call latency
	toolStart := time.Now()
	result, err := l.toolExecutor.ExecuteTools(ctx, state.toolName, toolContent)
	toolLatency := time.Since(toolStart).Milliseconds()
	toolCall.Latency = toolLatency
	toolCall.ToolOutput = result

	status := types.ToolStatusSuccess
	if err != nil {
		logger.WarnC(ctx, "tool execute failed", zap.String("tool", state.toolName), zap.Error(err))
		status = types.ToolStatusFailed
		result = fmt.Sprintf("%s execute failed, err: %v", state.toolName, err)
		toolCall.Error = err.Error()
	} else {
		logResult := result
		if len(logResult) > 400 {
			logResult = logResult[:400] + "..."
		}
		logger.InfoC(ctx, "tool execute succeed", zap.String("tool", state.toolName),
			zap.String("result", logResult), zap.Int("result length", len(result)))

		if len(result) > MaxToolResultLength {
			logger.WarnC(ctx, "tool result truncated due to excessive length",
				zap.String("tool", state.toolName),
				zap.Int("original_length", len(result)),
				zap.Int("truncated_length", MaxToolResultLength))
			result = result[:MaxToolResultLength] + "... (truncated due to excessive length)"
		}
	}
	toolCall.ResultStatus = string(status)

	l.request.Messages = append(l.request.Messages,
		types.Message{
			Role:    types.RoleAssistant,
			Content: state.fullContent.String(),
		},
		types.Message{
			Role: types.RoleUser,
			Content: []model.Content{
				{
					Type: model.ContTypeText,
					Text: fmt.Sprintf("[%s] Result:", state.toolName),
				}, {
					Type: model.ContTypeText,
					Text: result,
				}, {
					Type: model.ContTypeText,
					Text: fmt.Sprintf("Please summarize the key findings and/or code from the results above within the <think></think> tags. No need to summarize error messages. \nIf the search failed, don't say 'failed', describe this outcome as 'did not found relevant results' instead - MUST NOT using terms like 'failure', 'error', or 'unsuccessful' in your description. \nIn your summary, must include the name of the tool used and specify which tools you intend to use next. \nWhen appropriate, prioritize using these tools: %s", l.toolExecutor.GetAllTools()),
				},
			},
		},
	)

	l.updateToolStatus(state.toolName, status)
	chatLog.ProcessedPrompt = l.request.Messages
	chatLog.ToolCalls = append(chatLog.ToolCalls, toolCall)

	// sending tool call ending response to client page
	if err := l.sendStreamContent(flusher, state.response, types.StrFilterToolAnalyzing); err != nil {
		return err
	}
	for i := 0; i < 3; i++ {
		time.Sleep(100 * time.Millisecond)
		if err := l.sendStreamContent(flusher, state.response, "."); err != nil {
			return err
		}
	}
	if err := l.sendStreamContent(flusher, state.response, "\n"); err != nil {
		return err
	}

	// Recursive processing
	return l.handleStreamingWithTools(
		ctx,
		llmClient,
		flusher,
		chatLog,
		remainingDepth-1,
		idleTracker,
	)
}

// completeStreamResponse sends remaining content and updates statistics
func (l *ChatCompletionLogic) completeStreamResponse(
	flusher http.Flusher,
	chatLog *model.ChatLog,
	state *streamState,
) error {
	logger.InfoC(l.ctx, "starting to send remaining content before ending.")

	// Check if the entire response is invalid by verifying if we received any response data
	// Also check if the content is only empty (excluding newlines)
	fullContentStr := state.fullContent.String()
	trimmedContent := strings.ReplaceAll(fullContentStr, "\n", "")

	if state.response == nil || trimmedContent == "" {
		logger.WarnC(l.ctx, "detected invalid or empty response")

		// Send error response
		noContentErr := types.NewInvaildResponseContentError()
		l.responseHandler.sendSSEError(l.ctx, l.writer, noContentErr)
		chatLog.AddError(types.ErrApiError, noContentErr)
		return nil
	}

	if len(state.window) > 0 {
		if state.window[len(state.window)-1] == "[DONE]" {
			state.window = state.window[:len(state.window)-1]
		}

		endContent := strings.Join(state.window, "")

		if l.usage != nil {
			state.response.Usage = *l.usage
		} else {
			logger.WarnC(l.ctx, "usage is nil when content ending")
		}

		if err := l.sendStreamContent(flusher, state.response, endContent); err != nil {
			return err
		}

		if err := l.sendRawLine(flusher, "[DONE]"); err != nil {
			return err
		}
	}

	l.updateStreamStats(chatLog, state)

	return nil
}

// handleStreamError handles streaming errors with appropriate error responses
func (l *ChatCompletionLogic) handleStreamError(err error, chatLog *model.ChatLog) error {
	// Check if it's a context cancellation (client disconnect)
	if errors.Is(err, context.Canceled) || errors.Is(l.ctx.Err(), context.Canceled) {
		logger.WarnC(l.ctx, "Client disconnected (context canceled)", zap.Error(err))
		return nil
	}

	logger.ErrorC(l.ctx, "ChatLLMWithMessagesStreamRaw error", zap.Error(err))

	if l.isContextLengthError(err) {
		logger.ErrorC(l.ctx, "Input context too long", zap.Error(err))
		lengthErr := types.NewContextTooLongError()
		l.responseHandler.sendSSEError(l.ctx, l.writer, lengthErr)
		chatLog.AddError(types.ErrContextExceeded, lengthErr)
		return nil
	}

	l.responseHandler.sendSSEError(l.ctx, l.writer, err)
	chatLog.AddError(types.ErrApiError, err)
	return nil
}

// updateStreamStats updates chat log with streaming statistics
func (l *ChatCompletionLogic) updateStreamStats(chatLog *model.ChatLog, state *streamState) {
	endTime := time.Since(state.modelStart)
	logger.InfoC(l.ctx, "[last-token] stream end", zap.Duration("totalLatency", endTime))
	chatLog.Latency.MainModelLatency = endTime.Milliseconds()
	chatLog.ResponseContent = &types.ResponseContent{
		Content: state.fullContent.String(),
	}

	if l.usage != nil {
		chatLog.Usage = *l.usage
	} else {
		chatLog.Usage = l.responseHandler.calculateUsage(
			chatLog.Tokens.Processed.All,
			chatLog.ResponseContent.Content,
		)
		logger.InfoC(l.ctx, "calculated usage for streaming response")
	}

	logger.Info("prompt usage", zap.Any("usage", chatLog.Usage))
}

func (l *ChatCompletionLogic) sendRawLine(flusher http.Flusher, raw string) error {
	if !strings.HasPrefix(raw, "data: ") {
		raw = "data: " + raw
	}

	_, err := fmt.Fprintf(l.writer, "%s\n\n", raw)
	flusher.Flush()
	return err
}

func (l *ChatCompletionLogic) sendStreamContent(flusher http.Flusher, response *types.ChatCompletionResponse, content string) error {
	if response == nil {
		logger.WarnC(l.ctx, "response is nil, use default response", zap.String("method", "sendStreamContent"))
		response = &types.ChatCompletionResponse{}
	}

	response.Choices = []types.Choice{{
		Delta: types.Delta{
			Content: content,
		},
	}}
	jsonData, _ := json.Marshal(response)

	_, err := fmt.Fprintf(l.writer, "data: %s\n\n", jsonData)
	flusher.Flush()
	return err
}

// Helper methods

// getOrCreateRouterStrategy returns the cached router strategy instance or creates a new one
// This ensures that stateful strategies (like priority with round-robin) maintain their state across requests
func (l *ChatCompletionLogic) getOrCreateRouterStrategy() router.Strategy {
	// Try to get cached strategy from ServiceContext
	if cachedStrategy := l.svcCtx.GetRouterStrategy(); cachedStrategy != nil {
		if strategy, ok := cachedStrategy.(router.Strategy); ok {
			return strategy
		}
	}

	// Create new strategy if not cached
	if l.svcCtx.Config.Router == nil {
		return nil
	}

	strategy := router.NewRunner(*l.svcCtx.Config.Router)
	if strategy != nil {
		// Cache it for future requests
		l.svcCtx.SetRouterStrategy(strategy)
		logger.InfoC(l.ctx, "router strategy created and cached",
			zap.String("strategy", strategy.Name()))
	}

	return strategy
}

func (l *ChatCompletionLogic) updateToolStatus(toolName string, status types.ToolStatus) {
	if l.identity.RequestID == "" {
		logger.WarnC(l.ctx, "requestID is empty, skip updating tool status")
		return
	}
	toolStatusKey := types.ToolStatusRedisKeyPrefix + l.identity.RequestID

	if err := l.svcCtx.RedisClient.SetHashField(l.ctx, toolStatusKey, toolName, string(status), 5*time.Minute); err != nil {
		logger.ErrorC(l.ctx, "failed to update tool status in redis",
			zap.String("toolName", toolName),
			zap.String("status", string(status)),
			zap.Error(err))
	}

	logger.Info("Tool execute status updated", zap.String("tool", toolName),
		zap.String("execute status", string(status)))
}

// isContextLengthError checks if the error is due to context length exceeded
func (l *ChatCompletionLogic) isContextLengthError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "This model's maximum context length") ||
		strings.Contains(errMsg, "Input text is too long")
}

func (l *ChatCompletionLogic) callModelWithRetry(modelName string, params types.LLMRequestParams, idleTrackerOpt ...*timeout.IdleTracker) (types.ChatCompletionResponse, error) {
	nilResp := types.ChatCompletionResponse{}

	// Get config based on mode
	maxRetryCount, retryInterval, idleTimeout, totalIdleTimeout := l.getRetryConfig()

	// Use provided idle tracker if available, otherwise create a new one for this call
	var sharedTracker *timeout.IdleTracker
	if len(idleTrackerOpt) > 0 && idleTrackerOpt[0] != nil {
		sharedTracker = idleTrackerOpt[0]
	} else {
		sharedTracker = timeout.NewIdleTracker(totalIdleTimeout)
	}

	var lastErr error

	for attempt := 0; attempt <= maxRetryCount; attempt++ {
		logger.InfoC(l.ctx, "single-model retry: calling model",
			zap.String("model", modelName),
			zap.Int("attempt", attempt+1),
			zap.Int("maxRetries", maxRetryCount),
		)

		llmClient, err := client.NewLLMClient(l.svcCtx.Config.LLM, l.svcCtx.Config.LLMTimeout, modelName, l.headers)
		if err != nil {
			logger.WarnC(l.ctx, "single-model retry: failed to create llm client",
				zap.String("model", modelName), zap.Error(err))
			return nilResp, err
		}

		// Use the shared idle tracker instead of creating a new one
		timerCtx, timerCancel, idleTimer := timeout.NewIdleTimer(l.ctx, idleTimeout, sharedTracker)
		resp, err := llmClient.ChatLLMWithMessagesRaw(timerCtx, params, idleTimer)
		idleTimer.Stop()
		timerCancel()
		if err == nil {
			l.request.Model = modelName
			if l.writer != nil {
				l.writer.Header().Set(types.HeaderSelectLLm, modelName)
			}
			logger.InfoC(l.ctx, "single-model retry: model succeeded", zap.String("model", modelName))
			return resp, nil
		}

		lastErr = err
		retryable := isRetryableAPIError(err)
		logger.WarnC(l.ctx, "single-model retry: attempt failed",
			zap.String("model", modelName),
			zap.Bool("retryable", retryable),
			zap.Error(err),
		)

		if retryable && attempt < maxRetryCount {
			// Check if we have enough idle budget remaining for retry
			remainingIdleBudget := sharedTracker.Remaining()
			minRequiredBudget := retryInterval
			if remainingIdleBudget < minRequiredBudget {
				logger.WarnC(l.ctx, "single-model retry: insufficient idle budget for retry",
					zap.Duration("remainingIdleBudget", remainingIdleBudget),
					zap.Duration("minRequiredBudget", minRequiredBudget))
				break
			}
			time.Sleep(retryInterval)
			continue
		}

		break
	}

	return nilResp, lastErr
}

// callWithDegradation attempts models in l.orderedModels with idle timeout control.
// Retry the same model once (after 5s sleep) on timeout or 5xx errors; otherwise move to next.
func (l *ChatCompletionLogic) callWithDegradation(params types.LLMRequestParams, idleTracker *timeout.IdleTracker) (types.ChatCompletionResponse, error) {
	nilResp := types.ChatCompletionResponse{}
	if len(l.orderedModels) == 0 {
		return nilResp, fmt.Errorf("degradation: ordered models is empty")
	}
	// ensure unique order and place current at front
	ordered := make([]string, 0, len(l.orderedModels))
	seen := map[string]struct{}{}
	for _, m := range l.orderedModels {
		if m == "" {
			continue
		}
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		ordered = append(ordered, m)
	}
	if len(ordered) == 0 {
		return nilResp, fmt.Errorf("degradation: no valid model in order list")
	}

	var lastErr error
	for _, modelName := range ordered {
		logger.InfoC(l.ctx, "degradation: attempting model",
			zap.String("model", modelName),
		)

		resp, err := l.callModelWithRetry(modelName, params, idleTracker)
		if err == nil {
			logger.InfoC(l.ctx, "degradation: model succeeded", zap.String("model", modelName))
			return resp, nil
		}

		lastErr = err

		// If the error is context canceled, stop trying other models
		if errors.Is(err, context.Canceled) || errors.Is(l.ctx.Err(), context.Canceled) {
			logger.WarnC(l.ctx, "degradation: context canceled, stopping degradation",
				zap.String("model", modelName),
				zap.Error(err),
			)
			return nilResp, err
		}

		logger.WarnC(l.ctx, "degradation: model failed, moving to next",
			zap.String("model", modelName),
			zap.Error(err),
		)
	}
	return nilResp, lastErr
}

// isRetryableAPIError returns true when we should retry the same model: timeout/network/5xx
func isRetryableAPIError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	if apiErr, ok := err.(*types.APIError); ok {
		if apiErr.StatusCode >= http.StatusInternalServerError ||
			apiErr.Code == types.ErrCodeServerBusy ||
			apiErr.Code == types.ErrCodeModelServiceUnavailable {
			return true
		}
		return false
	}
	// fallback: retry once for non-APIError
	return true
}

// handleRawModeStream handles raw mode streaming by directly passing through LLM response
func (l *ChatCompletionLogic) handleRawModeStream(
	ctx context.Context,
	llmClient client.LLMInterface,
	flusher http.Flusher,
	chatLog *model.ChatLog,
	idleTracker *timeout.IdleTracker,
) error {
	logger.InfoC(ctx, "handling raw mode streaming - direct passthrough")

	// Direct call LLM streaming interface and pass through results
	modelStart := time.Now()
	firstTokenReceived := false
	var firstTokenTime time.Time
	var respStr strings.Builder // Accumulate full content for validation

	// Initialize accumulated response for function call format
	accumulatedResp := &types.ResponseContent{}
	toolCallsMap := make(map[int]*types.ToolCallInfo) // Map to accumulate tool calls by index

	// Use the provided shared idle tracker instead of creating a new one
	_, _, idleTimeout, _ := l.getRetryConfig()
	timerCtx, cancel, idleTimer := timeout.NewIdleTimer(ctx, idleTimeout, idleTracker)
	defer func() {
		idleTimer.Stop()
		cancel()
	}()

	err := llmClient.ChatLLMWithMessagesStreamRaw(timerCtx, l.request.LLMRequestParams, idleTimer, func(llmResp client.LLMResponse) error {
		// Handle response headers
		l.handleResonseHeaders(llmResp.Header, types.ResponseHeadersToForward, chatLog)

		// Direct pass through response line to client
		if llmResp.ResonseLine != "" {
			// Record first token time
			if !firstTokenReceived {
				firstTokenReceived = true
				firstTokenTime = time.Now()
				firstTokenLatency := firstTokenTime.Sub(modelStart)
				chatLog.Latency.FirstTokenLatency = firstTokenLatency.Milliseconds()
				logger.InfoC(ctx, "[first-token][raw mode] first token received, and response",
					zap.String("model", l.request.Model), zap.Duration("firstTokenLatency", firstTokenLatency))

				// 通知 idleTimer 已接收首token（新增）
				idleTimer.SetFirstTokenReceived()
			}

			// Extract usage information from streaming response
			_, usage, _ := l.responseHandler.extractStreamingData(llmResp.ResonseLine)
			if usage != nil {
				l.usage = usage
			}

			// Extract delta content for accumulated response
			l.responseHandler.extractSSEFunctionResp(llmResp.ResonseLine, accumulatedResp, toolCallsMap)

			// Check if the data line has valid content (not just "" or [DONE])
			if strings.HasPrefix(llmResp.ResonseLine, "data: ") {
				respLine := strings.TrimPrefix(llmResp.ResonseLine, "data: ")
				respLine = strings.TrimSpace(respLine)
				// Only accumulate if data is not empty string or [DONE]
				if respLine != "" && respLine != `""` && respLine != "[DONE]" {
					respStr.WriteString(respLine)
				}
			}

			if _, err := fmt.Fprintf(l.writer, "%s\n\n", llmResp.ResonseLine); err != nil {
				return err
			}
			flusher.Flush()
		}

		return nil
	})

	if err != nil {
		if l.isContextLengthError(err) {
			logger.ErrorC(ctx, "Input context too long in raw mode", zap.Error(err))
			lengthErr := types.NewContextTooLongError()
			l.responseHandler.sendSSEError(ctx, l.writer, lengthErr)
			chatLog.AddError(types.ErrContextExceeded, lengthErr)
			return nil
		}

		l.responseHandler.sendSSEError(ctx, l.writer, err)
		chatLog.AddError(types.ErrApiError, err)
		return nil
	}

	// Check if we received any valid content (same logic as completeStreamResponse)
	allRespStr := respStr.String()
	trimmedContent := strings.ReplaceAll(allRespStr, "\n", "")

	if trimmedContent == "" {
		logger.WarnC(ctx, "[raw mode] detected invalid or empty response")

		// Send error response
		noContentErr := types.NewInvaildResponseContentError()
		l.responseHandler.sendSSEError(ctx, l.writer, noContentErr)
		chatLog.AddError(types.ErrApiError, noContentErr)
		return nil
	}

	// Record statistics and total latency
	endTime := time.Now()
	totalLatency := endTime.Sub(modelStart)
	chatLog.Latency.MainModelLatency = totalLatency.Milliseconds()

	if firstTokenReceived {
		logger.InfoC(ctx, "[last-token][raw mode] last token received",
			zap.Duration("totalLatency", totalLatency))
	}

	// Record usage information
	if l.usage != nil {
		chatLog.Usage = *l.usage
		logger.InfoC(ctx, "[raw mode] prompt usage", zap.Any("usage", chatLog.Usage))
	} else {
		logger.InfoC(ctx, "[raw mode] no usage information available in streaming response")
	}

	logger.InfoC(ctx, "raw mode streaming completed",
		zap.Int64("modelLatency", chatLog.Latency.MainModelLatency))

	// Finalize and store accumulated response
	l.recordFuncionCallResponse(ctx, accumulatedResp, toolCallsMap, chatLog)

	return nil
}

// recordFuncionCallResponse converts tool calls map to slice and stores accumulated response in chatLog
func (l *ChatCompletionLogic) recordFuncionCallResponse(
	ctx context.Context,
	funCallResp *types.ResponseContent,
	toolCallsMap map[int]*types.ToolCallInfo,
	chatLog *model.ChatLog,
) {
	// Convert tool calls map to slice
	if len(toolCallsMap) > 0 {
		toolCallsSlice := make([]types.ToolCallInfo, 0, len(toolCallsMap))
		for i := 0; i < len(toolCallsMap); i++ {
			if tc, ok := toolCallsMap[i]; ok {
				toolCallsSlice = append(toolCallsSlice, *tc)
			}
		}
		funCallResp.ToolCalls = toolCallsSlice
	}

	// Store accumulated response in chatLog
	if funCallResp.Role != "" || funCallResp.Content != "" ||
		funCallResp.ReasoningContent != "" || len(funCallResp.ToolCalls) > 0 {
		chatLog.ResponseContent = funCallResp
		logger.InfoC(ctx, "[raw mode] recorded response content")
	}
}

// sanitizeHeaderValue removes CR/LF and control characters from header values and trims length.
// This mirrors the plugin behavior to prevent header injection/breakages.
func sanitizeHeaderValue(val string) string {
	if strings.TrimSpace(val) == "" {
		return ""
	}
	// remove CR/LF and other CTLs
	runes := make([]rune, 0, len(val))
	for _, r := range val {
		if r == '\r' || r == '\n' || r == 0x7f || r < 0x20 {
			continue
		}
		runes = append(runes, r)
	}
	out := strings.TrimSpace(string(runes))
	// limit length to avoid very long headers
	const maxLen = 128
	if len(out) > maxLen {
		out = out[:maxLen]
	}
	return out
}

func (l *ChatCompletionLogic) countTokensInMessages(messages []types.Message) int {
	if l.svcCtx.TokenCounter != nil {
		return l.svcCtx.TokenCounter.CountMessagesTokens(messages)
	}

	// Fallback to simple estimation
	totalText := ""
	for _, msg := range messages {
		totalText += msg.Role + ": " + utils.GetContentAsString(msg.Content) + "\n"
	}
	return tokenizer.EstimateTokens(totalText)
}
