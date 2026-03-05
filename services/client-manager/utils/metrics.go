package utils

import (
	"sync"
	"sync/atomic"
	"time"
)

// Global counters for metrics
var (
	requestCount uint64
	errorCount   uint64
	startupTime  time.Time
	timeMutex    sync.Mutex
)

/**
 * IncrementRequestCount increments the global request counter
 * @description
 * - Thread-safe increment of request counter
 * - Used by middleware to track total requests
 * - Supports concurrent access
 */
func IncrementRequestCount() {
	atomic.AddUint64(&requestCount, 1)
}

/**
 * GetRequestCount returns the total number of requests
 * @returns {uint64} Total request count
 * @description
 * - Thread-safe read of request counter
 * - Used for health checks and metrics
 * - Returns current atomic value
 */
func GetRequestCount() uint64 {
	return atomic.LoadUint64(&requestCount)
}

/**
 * IncrementErrorCount increments the global error counter
 * @description
 * - Thread-safe increment of error counter
 * - Used by middleware to track errors
 * - Supports concurrent access
 */
func IncrementErrorCount() {
	atomic.AddUint64(&errorCount, 1)
}

/**
 * GetErrorCount returns the total number of errors
 * @returns {uint64} Total error count
 * @description
 * - Thread-safe read of error counter
 * - Used for health checks and metrics
 * - Returns current atomic value
 */
func GetErrorCount() uint64 {
	return atomic.LoadUint64(&errorCount)
}

/**
 * SetStartupTime sets the application startup time
 * @param {time.Time} t - Startup time
 * @description
 * - Thread-safe setting of startup time
 * - Called during application initialization
 * - Used for uptime calculation
 */
func SetStartupTime(t time.Time) {
	timeMutex.Lock()
	defer timeMutex.Unlock()
	startupTime = t
}

/**
 * GetStartupTime returns the application startup time
 * @returns {time.Time} Startup time
 * @description
 * - Thread-safe read of startup time
 * - Used for uptime calculation
 * - Returns zero time if not set
 */
func GetStartupTime() time.Time {
	timeMutex.Lock()
	defer timeMutex.Unlock()
	return startupTime
}
