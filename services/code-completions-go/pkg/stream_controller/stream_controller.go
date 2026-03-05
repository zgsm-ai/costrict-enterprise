package stream_controller

import (
	"code-completion/pkg/completions"
	"code-completion/pkg/config"
	"code-completion/pkg/model"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// 全局流控管理器
var Controller *StreamController

// 流控管理器,对补全模型的访问做流控，防止补全模型失去响应
type StreamController struct {
	queues *QueueManager //请求等待队列管理（在等待调度到模型请求池）
	pools  *PoolManager  //模型请求池管理（正在调用模型的请求）
}

func NewStreamController() *StreamController {
	return &StreamController{
		queues: NewQueueManager(),
		pools:  NewPoolManager(),
	}
}

func (sc *StreamController) Init() {
	sc.pools.Init()

	var maintainInterval time.Duration
	maintainInterval = time.Duration(300) * time.Second // 默认清理间隔（秒）
	if config.Config.StreamController.MaintainInterval > 0 {
		maintainInterval = config.Config.StreamController.MaintainInterval
	}
	sc.StartMaintainRoutine(maintainInterval)

	zap.L().Info("Initialize queue configuration",
		zap.Duration("maintainInterval", maintainInterval))
}

/**
 * 处理V1接口版本的补全请求
 */
func (sc *StreamController) ProcessCompletionV1(ctx context.Context, input *completions.CompletionInput) *completions.CompletionResponse {
	var perf completions.CompletionPerformance
	perf.ReceiveTime = time.Now().Local()
	// 如果无法获取到clientID和completionID，拒掉
	if input.ClientID == "" || input.CompletionID == "" {
		return completions.CancelRequest(input.CompletionID, input.Model, &perf, model.StatusRejected, fmt.Errorf("missing client id or completion id"))
	}
	//	预选模型池
	pool := sc.pools.SelectIdlestPool(input.Model)
	if pool == nil {
		return completions.CancelRequest(input.CompletionID, input.Model, &perf, model.StatusBusy, fmt.Errorf("model pool busy, cancel request"))
	}
	input.Model = pool.cfg.ModelName

	//	上下文预处理
	c := completions.NewCompletionContext(ctx, &perf)
	rsp := input.Preprocess(c)
	if rsp != nil {
		return rsp
	}
	//	请求数据针对模型进行适应性改造
	handler := completions.NewCompletionHandler(pool.llm)
	para := handler.Adapt(input)

	// 将请求添加到客户端队列，获取包含响应通道的ClientRequest
	req := sc.queues.AddRequest(ctx, para, &perf)
	defer func() {
		sc.queues.RemoveRequest(req)
	}()
	return sc.pools.WaitDoRequest(req)
}

/**
 * ProcessCompletionV2 processes V2 interface version completion requests
 * @param {context.Context} ctx - Request context for controlling request lifecycle
 * @param {*model.CompletionParameter} para - Completion parameters containing request details and model information
 * @returns {*completions.CompletionResponse} Returns completion response with generated content or error information
 * @description
 * - Records performance metrics including receive time
 * - Adds request to queue manager for processing
 * - Automatically removes request from queue when function completes
 * - Waits for and executes the request through pool manager
 * - Handles V2 version completion requests with simplified flow compared to V1
 */
func (sc *StreamController) ProcessCompletionV2(ctx context.Context, para *model.CompletionParameter) *completions.CompletionResponse {
	var perf completions.CompletionPerformance
	perf.ReceiveTime = time.Now().Local()

	req := sc.queues.AddRequest(ctx, para, &perf)
	defer func() {
		sc.queues.RemoveRequest(req)
	}()
	return sc.pools.WaitDoRequest(req)
}

/**
 * ProcessCompletionOpenAI processes OpenAI format completion requests
 * @param {context.Context} ctx - Request context for controlling request lifecycle
 * @param {*model.CompletionRequest} r - OpenAI format completion request containing model parameters and prompt
 * @returns {*completions.CompletionResponse} Returns completion response with generated content or error information
 * @description
 * - Records performance metrics including receive time
 * - Finds the idlest model pool from all available pools
 * - Returns busy error if no available pool is found
 * - Creates completion handler with selected pool's LLM instance
 * - Creates completion context with performance tracking
 * - Directly handles the OpenAI format completion without queue management
 * - Designed for OpenAI API compatible request processing
 */
func (sc *StreamController) ProcessCompletionOpenAI(ctx context.Context, r *model.CompletionRequest) *completions.CompletionResponse {
	var perf completions.CompletionPerformance
	perf.ReceiveTime = time.Now().Local()

	pool := sc.pools.findIdlestPool(sc.pools.all)
	if pool == nil {
		return completions.CancelRequest("", r.Model, &perf, model.StatusBusy, fmt.Errorf("model pool busy, cancel request"))
	}
	handler := completions.NewCompletionHandler(pool.llm)
	c := completions.NewCompletionContext(ctx, &perf)
	return handler.HandleCompletionOpenAI(c, r)
}

/**
 * StartMaintainRoutine starts a goroutine for periodic maintenance operations
 * @param {time.Duration} interval - Time interval between maintenance operations
 * @description
 * - Creates a ticker with specified interval for periodic execution
 * - Runs cleanup operations on queues to remove stale requests
 * - Logs maintenance statistics and controller status
 * - Operates in background goroutine without blocking main thread
 * - Automatically stops ticker when goroutine exits
 * - Logs the start of maintenance routine with configured interval
 */
func (sc *StreamController) StartMaintainRoutine(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			sc.queues.Cleanup()
			zap.L().Info("StreamController maintain", zap.Any("stats", sc.GetStats()))
		}
	}()

	zap.L().Info("Start maintain routine", zap.Duration("interval", interval))
}

// 获取流控统计信息
func (sc *StreamController) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["queues"] = sc.queues.GetStats()
	stats["pools"] = sc.pools.GetStats()
	return stats
}

func (sc *StreamController) GetDetails() map[string]interface{} {
	details := make(map[string]interface{})
	details["queues"] = sc.queues.GetDetails()
	details["pools"] = sc.pools.GetDetails()
	return details
}
