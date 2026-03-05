package stream_controller

import (
	"code-completion/pkg/completions"
	"code-completion/pkg/config"
	"code-completion/pkg/metrics"
	"code-completion/pkg/model"
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// 每个模型建立一个请求池，管理正在调用该模型的补全请求
// 模型请求池
type ModelPool struct {
	llm      model.LLM
	cfg      *config.ModelConfig
	mutex    sync.RWMutex
	waits    chan *ClientRequest
	runnings map[string]*ClientRequest
}

// 模型请求池管理器
type PoolManager struct {
	pools map[string][]*ModelPool
	all   []*ModelPool
}

// 创建模型请求池管理器
func NewPoolManager() *PoolManager {
	return &PoolManager{
		pools: make(map[string][]*ModelPool),
		all:   make([]*ModelPool, 0),
	}
}

func (m *PoolManager) Init() {
	for i, cfg := range config.Config.Models {
		modelName := cfg.ModelName
		if modelName == "" {
			modelName = "default"
		}
		m.initPool(modelName, model.GetModel(i), &config.Config.Models[i])
	}
	if len(m.all) == 0 {
		zap.L().Error("Initialize model error, 'models' is missing",
			zap.Int("modelCount", len(config.Config.Models)))
		panic("config missing 'models'")
	}
}

// initPool 初始化模型请求池
func (m *PoolManager) initPool(model string, llm model.LLM, cfg *config.ModelConfig) *ModelPool {
	pool := &ModelPool{
		cfg:      cfg,
		llm:      llm,
		runnings: make(map[string]*ClientRequest),
		waits:    make(chan *ClientRequest, cfg.MaxConcurrent*2), // 缓冲区设为最大并发数的2倍
	}
	m.all = append(m.all, pool)

	// 启动MaxConcurrent个协程处理请求
	for i := 0; i < cfg.MaxConcurrent; i++ {
		go m.LoopDoRequest(pool)
	}

	// 将池添加到对应的模型名下
	if _, exists := m.pools[model]; !exists {
		m.pools[model] = make([]*ModelPool, 0)
	}
	m.pools[model] = append(m.pools[model], pool)

	// 为每个标签也添加相同的池
	for _, t := range cfg.Tags {
		if _, exists := m.pools[t]; !exists {
			m.pools[t] = make([]*ModelPool, 0)
		}
		m.pools[t] = append(m.pools[t], pool)
	}

	zap.L().Info("Initialize model pool",
		zap.String("model", model),
		zap.Int("maxConcurrent", cfg.MaxConcurrent))
	return pool
}

/**
* Find the model pool with the lowest load rate from a list of pools
* @param {[]*ModelPool} pools - List of model pools to search
* @returns {ModelPool} Returns the model pool with the lowest load rate
* @description
* - Iterates through the provided pools to find the one with the lowest load rate
* - Load rate is calculated as: active_requests / max_concurrent
* - If multiple pools have the same load rate, returns the first one found
* - If the list is empty, returns nil
* @example
* pool := manager.findLowestLoadPool(pools)
 */
func (m *PoolManager) findIdlestPool(pools []*ModelPool) *ModelPool {
	if len(pools) == 0 {
		return nil
	}

	lowestLoadRate := float64(1.0)
	var selectedPool *ModelPool
	for _, pool := range pools {
		pool.mutex.RLock()
		activeRequests := len(pool.runnings)
		maxConcurrent := pool.cfg.MaxConcurrent
		pool.mutex.RUnlock()

		if maxConcurrent <= 0 || activeRequests >= maxConcurrent {
			continue
		}
		// Calculate load rate (0.0 to 1.0, where 0.0 is idle and 1.0 is fully loaded)
		loadRate := float64(activeRequests) / float64(maxConcurrent)
		if loadRate < lowestLoadRate {
			lowestLoadRate = loadRate
			selectedPool = pool
		}
	}
	return selectedPool
}

func (m *PoolManager) SelectIdlestPool(modelName string) *ModelPool {
	var pool *ModelPool
	pools, exists := m.pools[modelName]
	if !exists || len(pools) == 0 {
		pool = m.findIdlestPool(m.all)
	} else {
		pool = m.findIdlestPool(pools)
	}
	return pool
}

// 等待模型池空闲处理请求
func (m *PoolManager) WaitDoRequest(req *ClientRequest) *completions.CompletionResponse {
	pool := m.SelectIdlestPool(req.Para.Model)
	if pool == nil {
		req.Canceled = true
		return completions.CancelRequest(req.Para.CompletionID, req.Para.Model, req.Perf, model.StatusBusy, fmt.Errorf("model pool busy, request rejected"))
	}
	req.Para.Model = pool.cfg.ModelName
	// 尝试将请求发送到ModelPool的waits通道，如果不能立即发送则失败
	select {
	case pool.waits <- req: // 成功将请求发送到waits通道
		// 等待请求处理完成,接收处理结果
		select {
		case rsp := <-req.rspChan:
			return rsp
		case <-req.ctx.Done():
			status := model.StatusTimeout
			if req.ctx.Err() == context.Canceled {
				status = model.StatusCanceled
			}
			req.Canceled = true
			return completions.CancelRequest(req.Para.CompletionID, req.Para.Model, req.Perf, status, req.ctx.Err())
		}
	default: // waits通道已满，无法立即发送请求
		zap.L().Debug("Model pool busy, failed to send request",
			zap.String("model", req.Para.Model),
			zap.String("clientID", req.Para.ClientID),
			zap.String("completionID", req.Para.CompletionID))
		req.Perf.QueueDuration = time.Since(req.Perf.EnqueueTime).Milliseconds()
		req.Canceled = true
		return completions.CancelRequest(req.Para.CompletionID, req.Para.Model, req.Perf, model.StatusBusy,
			fmt.Errorf("model pool busy, request rejected"))
	}
}

// LoopDoRequest 循环处理ModelPool的waits通道中的请求
func (m *PoolManager) LoopDoRequest(pool *ModelPool) {
	for {
		// 从waits通道获取请求
		req := <-pool.waits
		if req == nil || req.Canceled {
			continue
		}
		rsp := m.doRequest(pool, req)
		// 将结果发送回请求的响应通道
		select {
		case req.rspChan <- rsp:
		default:
			zap.L().Error("Failed to send response to client",
				zap.String("completionID", req.Para.CompletionID))
		}
	}
}

// 执行请求，调用补全模型
func (m *PoolManager) doRequest(pool *ModelPool, req *ClientRequest) *completions.CompletionResponse {
	req.Perf.QueueDuration = time.Since(req.Perf.EnqueueTime).Milliseconds()

	// 增加活跃请求计数
	pool.mutex.Lock()
	pool.runnings[req.Para.CompletionID] = req
	currentRequests := len(pool.runnings)
	pool.mutex.Unlock()

	metrics.UpdateCompletionConcurrentByModel(pool.cfg.ModelName, currentRequests)

	// 使用原有的补全处理器处理请求
	handler := completions.NewCompletionHandler(pool.llm)
	c := completions.NewCompletionContext(req.ctx, req.Perf)
	rsp := handler.CallLLM(c, req.Para)

	pool.mutex.Lock()
	delete(pool.runnings, req.Para.CompletionID)
	currentRequests = len(pool.runnings)
	pool.mutex.Unlock()

	metrics.UpdateCompletionConcurrentByModel(pool.cfg.ModelName, currentRequests)

	return rsp
}

// 获取统计信息
func (m *PoolManager) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["count"] = len(m.all)
	poolDetails := make([]map[string]interface{}, 0)
	for _, pool := range m.all {
		pool.mutex.RLock()
		poolInfo := map[string]interface{}{
			"name": pool.cfg.ModelName,
			"tags": pool.cfg.Tags,
			"requests": map[string]interface{}{
				"max_concurrent": pool.cfg.MaxConcurrent,
				"running":        len(pool.runnings),
				"waiting":        len(pool.waits),
			},
		}
		pool.mutex.RUnlock()
		poolDetails = append(poolDetails, poolInfo)
	}
	stats["pools"] = poolDetails
	return stats
}

func (m *PoolManager) GetDetails() map[string]interface{} {
	details := make(map[string]interface{})

	details["count"] = len(m.all)
	poolDetails := make([]map[string]interface{}, 0)
	for _, pool := range m.all {
		runnings := []map[string]interface{}{}
		pool.mutex.RLock()
		for _, req := range pool.runnings {
			runnings = append(runnings, req.GetSummary())
		}
		poolInfo := map[string]interface{}{
			"name": pool.cfg.ModelName,
			"tags": pool.cfg.Tags,
			"requests": map[string]interface{}{
				"max_concurrent": pool.cfg.MaxConcurrent,
				"running":        len(pool.runnings),
				"waiting":        len(pool.waits),
				"runnings":       runnings,
			},
		}
		pool.mutex.RUnlock()
		poolDetails = append(poolDetails, poolInfo)
	}
	details["pools"] = poolDetails
	return details
}
