package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

const (
	// Redis键前缀
	fileStatusPrefix = "file:status:"
	// 请求ID键前缀
	requestIdPrefix = "request:id:"
)

// StatusManager 文件状态管理器
type StatusManager struct {
	client            *redis.Client
	defaultExpiration time.Duration
}

// NewStatusManager 创建新的状态管理器
func NewStatusManager(client *redis.Client) *StatusManager {
	return &StatusManager{
		client:            client,
		defaultExpiration: 24 * time.Hour, // 默认24小时，保持向后兼容
	}
}

// NewStatusManagerWithExpiration 创建带有自定义过期时间的状态管理器
func NewStatusManagerWithExpiration(client *redis.Client, expiration time.Duration) *StatusManager {
	return &StatusManager{
		client:            client,
		defaultExpiration: expiration,
	}
}

// SetFileStatusByRequestId 通过RequestId设置文件处理状态到Redis
func (sm *StatusManager) SetFileStatusByRequestId(ctx context.Context, requestId string, status *types.FileStatusResponseData) error {
	key := sm.generateRequestKey(requestId)
	// 序列化状态数据
	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status data: %w", err)
	}

	// 设置到Redis，带过期时间
	err = sm.client.Set(ctx, key, data, sm.defaultExpiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set status in redis: %w", err)
	}

	return nil
}

// GetFileStatus 从Redis获取文件处理状态
func (sm *StatusManager) GetFileStatus(ctx context.Context, requestId string) (*types.FileStatusResponseData, error) {
	key := sm.generateRequestKey(requestId)
	// 从Redis获取数据
	data, err := sm.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// 键不存在，返回nil
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get status from redis: %w", err)
	}

	// 反序列化状态数据
	var status types.FileStatusResponseData
	err = json.Unmarshal([]byte(data), &status)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal status data: %w", err)
	}

	return &status, nil
}

func (sm *StatusManager) UpdateFileStatus(ctx context.Context, requestId string, updateFn func(*types.FileStatusResponseData)) error {
	key := sm.generateRequestKey(requestId)
	// 从Redis获取数据
	data, err := sm.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// 键不存在，创建新的状态
			currentStatus := &types.FileStatusResponseData{
				Process:       "pending",
				TotalProgress: 0,
			}
			updateFn(currentStatus)
			return sm.SetFileStatusByRequestId(ctx, requestId, currentStatus)
		}
		return fmt.Errorf("failed to get status from redis: %w", err)
	}

	// 反序列化状态数据
	var currentStatus types.FileStatusResponseData
	err = json.Unmarshal([]byte(data), &currentStatus)
	if err != nil {
		return fmt.Errorf("failed to unmarshal status data: %w", err)
	}

	// 应用更新函数
	updateFn(&currentStatus)

	// 保存更新后的状态
	return sm.SetFileStatusByRequestId(ctx, requestId, &currentStatus)
}

// DeleteFileStatus 删除文件处理状态
func (sm *StatusManager) DeleteFileStatus(ctx context.Context, clientID, codebasePath, codebaseName string) error {
	key := sm.generateKey(clientID, codebasePath, codebaseName)
	return sm.client.Del(ctx, key).Err()
}

// generateKey 生成Redis键
func (sm *StatusManager) generateKey(clientID, codebasePath, codebaseName string) string {
	// 使用clientID、codebasePath和codebaseName组合生成唯一键
	return fmt.Sprintf("%s%s:%s:%s", fileStatusPrefix, clientID, codebasePath, codebaseName)
}

// generateRequestKey 生成基于RequestId的Redis键
func (sm *StatusManager) generateRequestKey(requestId string) string {
	return fmt.Sprintf("%s%s", requestIdPrefix, requestId)
}

// SetExpiration 设置自定义过期时间
func (sm *StatusManager) SetExpiration(ctx context.Context, clientID, codebasePath, codebaseName string, expiration time.Duration) error {
	key := sm.generateKey(clientID, codebasePath, codebaseName)
	return sm.client.Expire(ctx, key, expiration).Err()
}

// CheckConnection 检查Redis连接
func (sm *StatusManager) CheckConnection(ctx context.Context) error {
	return sm.client.Ping(ctx).Err()
}

// ScanRunningTasks 扫描运行中的任务
func (sm *StatusManager) ScanRunningTasks(ctx context.Context) ([]types.RunningTaskInfo, error) {
	var runningTasks []types.RunningTaskInfo

	// 使用SCAN命令避免阻塞
	iter := sm.client.Scan(ctx, 0, "request:id:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		// 获取任务状态数据
		data, err := sm.client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue // 键可能已过期
			}
			return nil, fmt.Errorf("failed to get task data for key %s: %w", key, err)
		}

		// 解析任务状态
		var status types.FileStatusResponseData
		if err := json.Unmarshal([]byte(data), &status); err != nil {
			continue // 跳过格式错误的数据
		}

		// 过滤运行中的任务状态
		if sm.isRunningStatus(status.Process) {
			taskInfo, err := sm.parseTaskInfo(key, status)
			if err != nil {
				continue
			}
			runningTasks = append(runningTasks, taskInfo)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("redis scan error: %w", err)
	}

	// 按开始时间排序
	sort.Slice(runningTasks, func(i, j int) bool {
		return runningTasks[i].StartTime.After(runningTasks[j].StartTime)
	})

	return runningTasks, nil
}

// isRunningStatus 检查是否为运行中的状态
func (sm *StatusManager) isRunningStatus(status string) bool {
	return status == "pending" || status == "processing" || status == "running"
}

// parseTaskInfo 解析任务信息
func (sm *StatusManager) parseTaskInfo(key string, status types.FileStatusResponseData) (types.RunningTaskInfo, error) {
	// 从key中提取任务ID
	taskId := strings.TrimPrefix(key, "request:id:")

	// 解析任务状态数据中的时间信息
	var startTime, lastUpdateTime time.Time
	var estimatedCompletionTime *time.Time

	// 设置当前时间为最后更新时间
	lastUpdateTime = time.Now()

	// 根据进度估算开始时间
	if status.TotalProgress > 0 {
		startTime = lastUpdateTime.Add(-time.Duration(status.TotalProgress) * time.Minute)
	} else {
		startTime = lastUpdateTime
	}

	// 如果进度大于0且小于100，估算完成时间
	if status.TotalProgress > 0 && status.TotalProgress < 100 {
		estimatedTime := lastUpdateTime.Add(time.Duration((100 - status.TotalProgress)) * time.Minute)
		estimatedCompletionTime = &estimatedTime
	}

	// 尝试从文件列表中提取客户端ID
	var clientId string
	if len(status.FileList) > 0 {
		// 这里可以根据实际业务逻辑提取客户端ID
		// 暂时使用空字符串，后续可以根据需要扩展
		clientId = ""
	}

	return types.RunningTaskInfo{
		TaskId:                  taskId,
		ClientId:                clientId,
		Status:                  status.Process,
		Process:                 status.Process,
		TotalProgress:           status.TotalProgress,
		StartTime:               startTime,
		LastUpdateTime:          lastUpdateTime,
		EstimatedCompletionTime: estimatedCompletionTime,
		FileList:                status.FileList,
	}, nil
}

// ResetPendingAndProcessingTasksToFailed 将所有pending和processing任务状态重置为failed
func (sm *StatusManager) ResetPendingAndProcessingTasksToFailed(ctx context.Context) error {
	// 使用SCAN命令避免阻塞
	iter := sm.client.Scan(ctx, 0, "request:id:*", 0).Iterator()

	var updatedCount int
	for iter.Next(ctx) {
		key := iter.Val()

		// 获取任务状态数据
		data, err := sm.client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue // 键可能已过期
			}
			return fmt.Errorf("failed to get task data for key %s: %w", key, err)
		}

		// 解析任务状态
		var status types.FileStatusResponseData
		if err := json.Unmarshal([]byte(data), &status); err != nil {
			continue // 跳过格式错误的数据
		}

		// 检查是否为pending或processing状态
		if status.Process == "pending" || status.Process == "processing" {
			// 更新状态为failed
			status.Process = "failed"

			// 序列化更新后的状态数据
			updatedData, err := json.Marshal(status)
			if err != nil {
				logx.Errorf("failed to marshal updated status data for key %s: %v", key, err)
				continue
			}

			// 更新Redis中的状态
			if err := sm.client.Set(ctx, key, updatedData, sm.defaultExpiration).Err(); err != nil {
				logx.Errorf("failed to update status to failed for key %s: %v", key, err)
				continue
			}

			updatedCount++
			logx.Infof("Reset task %s from %s to failed", strings.TrimPrefix(key, "request:id:"), status.Process)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("redis scan error: %w", err)
	}

	logx.Infof("Successfully reset %d pending/processing tasks to failed", updatedCount)
	return nil
}

// ScanCompletedTasks 扫描已完成的任务
func (sm *StatusManager) ScanCompletedTasks(ctx context.Context) ([]types.CompletedTaskInfo, error) {
	var completedTasks []types.CompletedTaskInfo

	// 使用SCAN命令避免阻塞
	iter := sm.client.Scan(ctx, 0, "request:id:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		// 获取任务状态数据
		data, err := sm.client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue // 键可能已过期
			}
			return nil, fmt.Errorf("failed to get task data for key %s: %w", key, err)
		}

		// 解析任务状态
		var status types.FileStatusResponseData
		if err := json.Unmarshal([]byte(data), &status); err != nil {
			continue // 跳过格式错误的数据
		}

		// 过滤已完成的任务状态
		if sm.isCompletedStatus(status.Process) {
			taskInfo, err := sm.parseCompletedTaskInfo(key, status)
			if err != nil {
				continue
			}
			completedTasks = append(completedTasks, taskInfo)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("redis scan error: %w", err)
	}

	// 按完成时间排序（最新的在前）
	sort.Slice(completedTasks, func(i, j int) bool {
		return completedTasks[i].CompletedTime.After(completedTasks[j].CompletedTime)
	})

	return completedTasks, nil
}

// isCompletedStatus 检查是否为已完成的状态
func (sm *StatusManager) isCompletedStatus(status string) bool {
	return status == "completed"
}

// parseCompletedTaskInfo 解析已完成任务信息
func (sm *StatusManager) parseCompletedTaskInfo(key string, status types.FileStatusResponseData) (types.CompletedTaskInfo, error) {
	// 从key中提取任务ID
	taskId := strings.TrimPrefix(key, "request:id:")

	// 解析任务状态数据中的时间信息
	var completedTime time.Time

	// 设置当前时间为完成时间
	completedTime = time.Now()

	// 尝试从文件列表中提取客户端ID
	var clientId string
	var fileCount int
	var successCount, failedCount int

	if len(status.FileList) > 0 {
		// 统计文件状态
		for _, file := range status.FileList {
			fileCount++
			if file.Status == "complete" {
				successCount++
			} else if file.Status == "failed" {
				failedCount++
			}
		}
	}

	// 计算成功率
	var successRate float64
	if fileCount > 0 {
		successRate = float64(successCount) / float64(fileCount) * 100
	}

	return types.CompletedTaskInfo{
		TaskId:        taskId,
		ClientId:      clientId,
		Status:        status.Process,
		Process:       status.Process,
		TotalProgress: status.TotalProgress,
		CompletedTime: completedTime,
		FileCount:     fileCount,
		SuccessCount:  successCount,
		FailedCount:   failedCount,
		SuccessRate:   successRate,
		FileList:      status.FileList,
	}, nil
}

// ScanFailedTasks 扫描失败的任务
func (sm *StatusManager) ScanFailedTasks(ctx context.Context) ([]types.CompletedTaskInfo, error) {
	var failedTasks []types.CompletedTaskInfo

	// 使用SCAN命令避免阻塞
	iter := sm.client.Scan(ctx, 0, "request:id:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		// 获取任务状态数据
		data, err := sm.client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue // 键可能已过期
			}
			return nil, fmt.Errorf("failed to get task data for key %s: %w", key, err)
		}

		// 解析任务状态
		var status types.FileStatusResponseData
		if err := json.Unmarshal([]byte(data), &status); err != nil {
			continue // 跳过格式错误的数据
		}

		// 过滤失败的任务状态
		if sm.isFailedStatus(status.Process) {
			taskInfo, err := sm.parseCompletedTaskInfo(key, status)
			if err != nil {
				continue
			}
			failedTasks = append(failedTasks, taskInfo)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("redis scan error: %w", err)
	}

	// 按完成时间排序（最新的在前）
	sort.Slice(failedTasks, func(i, j int) bool {
		return failedTasks[i].CompletedTime.After(failedTasks[j].CompletedTime)
	})

	return failedTasks, nil
}

// isFailedStatus 检查是否为失败的状态
func (sm *StatusManager) isFailedStatus(status string) bool {
	return status == "failed"
}
