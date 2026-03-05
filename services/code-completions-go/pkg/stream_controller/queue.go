package stream_controller

import (
	"code-completion/pkg/completions"
	"code-completion/pkg/config"
	"code-completion/pkg/metrics"
	"code-completion/pkg/model"
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

//
//	等待队列: 所有来自客户端的请求先排队，等待调度到模型请求池
//

// 客户端
type CompletionClient struct {
	ClientID   string
	Latest     *ClientRequest
	LatestTime time.Time
}

// 等待队列管理器
type QueueManager struct {
	clients  map[string]*CompletionClient
	requests map[string]*ClientRequest
	mutex    sync.RWMutex
}

// 创建等待队列管理器
func NewQueueManager() *QueueManager {
	return &QueueManager{
		clients:  make(map[string]*CompletionClient),
		requests: make(map[string]*ClientRequest),
	}
}

// 添加请求到等待队列
func (m *QueueManager) AddRequest(ctx context.Context, para *model.CompletionParameter, perf *completions.CompletionPerformance) *ClientRequest {
	reqCtx, cancel := context.WithTimeout(ctx, config.Config.StreamController.CompletionTimeout)
	req := &ClientRequest{
		Para:     para,
		Perf:     perf,
		Canceled: false,
		ctx:      reqCtx,
		cancel:   cancel,
		rspChan:  make(chan *completions.CompletionResponse, 1),
	}
	req.Perf.EnqueueTime = time.Now().Local()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	client, exists := m.clients[para.ClientID]
	if !exists {
		client = &CompletionClient{
			ClientID: para.ClientID,
		}
		m.clients[para.ClientID] = client
	}
	client.LatestTime = req.Perf.ReceiveTime
	if client.Latest != nil {
		m.cancelRequest(client.Latest)
		client.Latest = nil
	}
	client.Latest = req

	m.requests[para.ClientID+para.CompletionID] = req
	metrics.UpdateCompletionConcurrent(len(m.requests))
	return req
}

func (m *QueueManager) RemoveRequest(req *ClientRequest) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.requests, req.Para.ClientID+req.Para.CompletionID)
	metrics.UpdateCompletionConcurrent(len(m.requests))

	queue, exists := m.clients[req.Para.ClientID]
	if !exists {
		return
	}
	if queue.Latest == req {
		queue.Latest = nil
	}
}

// 取消现有请求
func (m *QueueManager) cancelRequest(req *ClientRequest) {
	zap.L().Debug("Cancel request",
		zap.String("clientID", req.Para.ClientID),
		zap.String("completionID", req.Para.CompletionID))
	if req.cancel != nil {
		req.cancel()
	}
	req.Canceled = true
}

// 清理过期的队列
func (m *QueueManager) Cleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 清理长时间没有活动的客户端
	currentTime := time.Now()
	for _, client := range m.clients {
		if currentTime.Sub(client.LatestTime) > config.Config.StreamController.CleanOlderThan {
			delete(m.clients, client.ClientID)
			zap.L().Info("Removed client", zap.String("clientID", client.ClientID),
				zap.Time("latestTime", client.LatestTime))
		}
	}
}

// 获取统计信息
func (m *QueueManager) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	activatedClient := 0
	for _, client := range m.clients {
		if client.Latest != nil {
			activatedClient++
		}
	}
	stats := make(map[string]interface{})
	stats["requests"] = map[string]interface{}{
		"total": len(m.requests),
	}
	stats["clients"] = map[string]interface{}{
		"activated": activatedClient,
		"total":     len(m.clients),
	}

	return stats
}

// 获取统计信息
func (m *QueueManager) GetDetails() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	activatedClient := 0
	clients := []map[string]interface{}{}
	for _, client := range m.clients {
		if client.Latest != nil {
			activatedClient++
			clients = append(clients, map[string]interface{}{
				"client_id":   client.ClientID,
				"latest":      client.Latest.GetSummary(),
				"latest_time": client.LatestTime,
			})
		} else {
			clients = append(clients, map[string]interface{}{
				"client_id":   client.ClientID,
				"latest_time": client.LatestTime,
			})
		}
	}
	requests := []map[string]interface{}{}
	for _, req := range m.requests {
		requests = append(requests, req.GetDetails())
	}

	return map[string]interface{}{
		"requests": map[string]interface{}{
			"total":   len(m.requests),
			"details": requests,
		},
		"clients": map[string]interface{}{
			"activated": activatedClient,
			"total":     len(m.clients),
			"details":   clients,
		},
	}
}
