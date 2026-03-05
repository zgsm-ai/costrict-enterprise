package timeout

import (
	"context"
	"sync"
	"time"

	"github.com/zgsm-ai/chat-rag/internal/logger"
	"go.uber.org/zap"
)

// IdleTimeoutReason represents the reason for idle timeout
type IdleTimeoutReason string

const (
	IdleTimeoutReasonPerIdle IdleTimeoutReason = "per_idle"
	IdleTimeoutReasonTotal   IdleTimeoutReason = "total"
)

// IdleTracker maintains the total idle budget across retries/degradations
type IdleTracker struct {
	mu              sync.Mutex
	initialBudget   time.Duration
	remainingBudget time.Duration
	lastResetTime   time.Time
}

// NewIdleTracker creates a new IdleTracker with the specified total budget
func NewIdleTracker(totalBudget time.Duration) *IdleTracker {
	return &IdleTracker{
		initialBudget:   totalBudget,
		remainingBudget: totalBudget,
		lastResetTime:   time.Now(),
	}
}

// Remaining returns the remaining total idle budget
func (t *IdleTracker) Remaining() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.remainingBudget
}

// Consume reduces the remaining budget by the specified duration
func (t *IdleTracker) Consume(duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.remainingBudget -= duration
	if t.remainingBudget < 0 {
		t.remainingBudget = 0
	}
}

// Reset resets both the single idle timer and the total idle tracker
func (t *IdleTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	// Reset total budget to initial value
	t.remainingBudget = t.initialBudget
	t.lastResetTime = time.Now()
}

// IdleTimer manages idle timeout for a single request attempt
type IdleTimer struct {
	ctx           context.Context
	cancel        context.CancelFunc
	perIdle       time.Duration
	tracker       *IdleTracker
	timer         *time.Timer
	mu            sync.Mutex
	reason        IdleTimeoutReason
	stopped       bool
	lastResetTime time.Time
	idleStartTime time.Time
	resetCount    int64
	generation    int64 // Generation counter to detect stale timeout events

	// 首token接收状态相关字段（新增）
	firstTokenReceived bool       // 标记是否已接收首token
	firstTokenMu       sync.Mutex // 保护 firstTokenReceived

	timedOut bool // 标记是否已触发超时
}

// NewIdleTimer creates a new IdleTimer with the specified per-idle timeout and tracker
// Returns the context, cancel function, and the timer instance
func NewIdleTimer(parentCtx context.Context, perIdle time.Duration, tracker *IdleTracker) (context.Context, context.CancelFunc, *IdleTimer) {
	ctx, cancel := context.WithCancel(parentCtx)

	it := &IdleTimer{
		ctx:           ctx,
		cancel:        cancel,
		perIdle:       perIdle,
		tracker:       tracker,
		reason:        IdleTimeoutReasonPerIdle,
		lastResetTime: time.Now(),
		idleStartTime: time.Now(),
		resetCount:    0,
	}

	// Check if total budget is already exhausted
	if tracker.Remaining() <= 0 {
		it.reason = IdleTimeoutReasonTotal
		cancel()
		logger.Info("IdleTimer: total budget exhausted at creation",
			zap.Duration("remaining", tracker.Remaining()))
		return ctx, cancel, it
	}

	// Always use perIdle as the initial window, regardless of remaining budget
	it.timer = time.NewTimer(perIdle)

	// Start watching in a goroutine
	go it.watch()

	logger.Info("IdleTimer created",
		zap.Duration("perIdle", perIdle),
		zap.Duration("totalRemaining", tracker.Remaining()))

	return ctx, cancel, it
}

// watch monitors the timer and contexts
func (it *IdleTimer) watch() {
	for {
		select {
		case <-it.ctx.Done():
			// Parent context cancelled
			it.mu.Lock()
			if it.timer != nil {
				it.timer.Stop()
			}
			it.mu.Unlock()
			return

		case <-it.timer.C:
			// Timer expired - capture generation immediately after reading from channel
			it.mu.Lock()
			capturedGen := it.generation
			stopped := it.stopped
			it.mu.Unlock()

			if stopped {
				return
			}

			// Handle timeout with the captured generation
			// If it's a real timeout (not stale), handleTimeout will cancel the context
			// and we'll exit via the ctx.Done() case in the next iteration
			it.handleTimeout(capturedGen)
		}
	}
}

// handleTimeout is called when the timer expires
// expectedGen is the generation captured immediately after reading from timer.C
func (it *IdleTimer) handleTimeout(expectedGen int64) {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.stopped {
		return
	}

	// Check if the generation has changed (meaning Reset() was called after timer fired)
	if it.generation != expectedGen {
		logger.Debug("IdleTimer: ignoring stale timeout event",
			zap.Int64("expectedGen", expectedGen),
			zap.Int64("currentGen", it.generation))
		return
	}

	// Calculate actual idle duration since last reset
	actualIdleDuration := time.Since(it.idleStartTime)

	// Consume the perIdle duration from total budget
	it.tracker.Consume(it.perIdle)

	remaining := it.tracker.Remaining()

	// Check if total budget is exhausted
	if remaining <= 0 {
		it.reason = IdleTimeoutReasonTotal
		logger.Warn("IdleTimer: total budget exhausted",
			zap.Duration("consumed", it.tracker.initialBudget-remaining))
	} else {
		it.reason = IdleTimeoutReasonPerIdle
	}

	// 读取首token状态
	it.firstTokenMu.Lock()
	firstTokenReceived := it.firstTokenReceived
	it.firstTokenMu.Unlock()

	// 核心逻辑：首token后（包括总预算耗尽），只记录日志，不取消上下文
	if firstTokenReceived {
		logger.Warn("IdleTimer: post-first-token timeout (logging only)",
			zap.Duration("perIdle", it.perIdle),
			zap.Duration("actualIdleDuration", actualIdleDuration),
			zap.Duration("remainingBudget", remaining),
			zap.Int64("resetCount", it.resetCount),
			zap.String("reason", string(it.reason)))
		return // 不取消上下文，让请求继续
	}

	// 首token前：取消上下文，返回错误
	logger.Warn("IdleTimer: timeout triggered before first token",
		zap.Duration("perIdle", it.perIdle),
		zap.Duration("actualIdleDuration", actualIdleDuration),
		zap.Duration("remainingBudget", remaining),
		zap.Int64("resetCount", it.resetCount),
		zap.String("reason", string(it.reason)))
	it.timedOut = true
	it.cancel()
}

// Reset resets the idle timer when data is received
func (it *IdleTimer) Reset() {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.stopped {
		return
	}

	// Increment generation to invalidate any pending timeout events
	it.generation++

	// Reset the tracker's total budget
	it.tracker.Reset()

	// Reset the timer to perIdle duration
	// According to Go documentation, we must drain the channel if Stop() returns false
	if it.timer != nil {
		// Stop the timer and check if it was already expired
		if !it.timer.Stop() {
			// Timer already expired, drain the channel to avoid blocking
			// Use select with default to avoid blocking if watch() goroutine already consumed it
			select {
			case <-it.timer.C:
			default:
			}
		}
		it.timer.Reset(it.perIdle)
	}

	it.lastResetTime = time.Now()
	it.idleStartTime = time.Now()
	it.resetCount++
	it.timedOut = false

	logger.Debug("IdleTimer reset",
		zap.Duration("perIdle", it.perIdle),
		zap.Duration("remainingBudget", it.tracker.Remaining()),
		zap.Int64("resetCount", it.resetCount),
		zap.Int64("generation", it.generation))
}

// Reason returns the reason for the timeout
func (it *IdleTimer) Reason() IdleTimeoutReason {
	it.mu.Lock()
	defer it.mu.Unlock()
	return it.reason
}

// Stop stops the timer and prevents further timeouts
func (it *IdleTimer) Stop() {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.stopped {
		return
	}

	it.stopped = true
	if it.timer != nil {
		it.timer.Stop()
	}

	logger.Debug("IdleTimer stopped",
		zap.Int64("resetCount", it.resetCount),
		zap.Duration("remainingBudget", it.tracker.Remaining()))
}

// SetFirstTokenReceived marks that the first token has been received
// This should only be called in streaming scenarios
func (it *IdleTimer) SetFirstTokenReceived() {
	it.firstTokenMu.Lock()
	defer it.firstTokenMu.Unlock()
	it.firstTokenReceived = true
	logger.Info("IdleTimer: first token received, subsequent timeouts will only log warnings")
}

// GetResetCount returns the number of times Reset was called
func (it *IdleTimer) GetResetCount() int64 {
	it.mu.Lock()
	defer it.mu.Unlock()
	return it.resetCount
}

// IsTimedOut returns true if the timer has triggered a timeout
func (it *IdleTimer) IsTimedOut() bool {
	it.mu.Lock()
	defer it.mu.Unlock()
	return it.timedOut
}
