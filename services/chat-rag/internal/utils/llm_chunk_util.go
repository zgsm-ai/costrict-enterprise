package utils

import (
	"math"
	"sort"
	"sync"
	"time"
)

// ChunkStats 用于统计 SSE chunk 的时间指标
type ChunkStats struct {
	mu sync.Mutex

	intervals []float32 // 存储每个 chunk 间隔时间（毫秒）
	lastTime  time.Time // 上一个 chunk 到达时间

	closed bool // 是否已结束（End 或 Stop）
}

// NewChunkStats 创建一个新的 ChunkStats 实例
func NewChunkStats(capacity int) *ChunkStats {
	if capacity <= 0 {
		capacity = 21333 // 默认容量（32K * 2/3）
	}
	return &ChunkStats{
		intervals: make([]float32, 0, capacity),
	}
}

// OnChunkArrived 记录一个 chunk 到达（自动开始统计）
func (cs *ChunkStats) OnChunkArrived() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.closed {
		return // 已结束，忽略
	}

	now := time.Now()
	if cs.lastTime.IsZero() {
		// 首次调用，自动开始统计
		cs.lastTime = now
		return
	}

	interval := float32(now.Sub(cs.lastTime).Milliseconds())
	cs.intervals = append(cs.intervals, interval)
	cs.lastTime = now
}

// OnChunkArrivedWithInterval 记录一个 chunk 到达，使用传入的时间间隔（毫秒）
func (cs *ChunkStats) OnChunkArrivedWithInterval(intervalMs float32) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.closed {
		return // 已结束，忽略
	}

	if cs.lastTime.IsZero() {
		// 首次调用，自动开始统计
		cs.lastTime = time.Now()
		return
	}

	cs.intervals = append(cs.intervals, intervalMs)
}

// End 正常结束统计，返回统计结果并释放内存
func (cs *ChunkStats) End() *ChunkStatInfo {
	return cs.finalize(false)
}

// Stop 异常结束统计，返回统计结果并释放内存
func (cs *ChunkStats) Stop() *ChunkStatInfo {
	return cs.finalize(true)
}

// finalize 内部方法，执行结束逻辑（幂等）
func (cs *ChunkStats) finalize(isError bool) *ChunkStatInfo {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.closed {
		return nil // 已结束，返回 nil（幂等）
	}

	cs.closed = true

	if len(cs.intervals) == 0 {
		// 无数据，释放内存
		cs.intervals = nil
		return nil
	}

	// 计算统计指标
	stats := cs.calculateStats(isError)

	// 释放内存
	cs.intervals = nil

	return stats
}

// calculateStats 计算统计指标
func (cs *ChunkStats) calculateStats(isError bool) *ChunkStatInfo {
	n := len(cs.intervals)
	if n == 0 {
		return nil
	}

	// 创建副本用于排序（计算百分位数）
	sorted := make([]float32, n)
	copy(sorted, cs.intervals)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// 计算基本统计量
	var sum float32
	min := sorted[0]
	max := sorted[n-1]

	for _, v := range cs.intervals {
		sum += v
	}
	mean := float32(sum) / float32(n)

	// 计算方差
	var variance float64
	for _, v := range cs.intervals {
		diff := float64(v) - float64(mean)
		variance += diff * diff
	}
	variance /= float64(n)

	// 计算百分位数
	p50 := sorted[n*50/100]
	p95 := sorted[n*95/100]
	p99 := sorted[n*99/100]

	return &ChunkStatInfo{
		Count:    n,
		Mean:     mean,
		Min:      min,
		Max:      max,
		Variance: variance,
		StdDev:   math.Sqrt(variance),
		P50:      p50,
		P95:      p95,
		P99:      p99,
		IsError:  isError,
	}
}

// ChunkStatInfo 统计结果
type ChunkStatInfo struct {
	Count    int     // chunk 数量
	Mean     float32 // 平均间隔时间（毫秒）
	Min      float32 // 最小间隔时间（毫秒）
	Max      float32 // 最大间隔时间（毫秒）
	Variance float64 // 方差
	StdDev   float64 // 标准差
	P50      float32 // 中位数（毫秒）
	P95      float32 // P95（毫秒）
	P99      float32 // P99（毫秒）
	IsError  bool    // 是否异常结束
}
