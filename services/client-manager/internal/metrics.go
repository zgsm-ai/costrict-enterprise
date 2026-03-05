package internal

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"

	"github.com/zgsm-ai/client-manager/utils"
)

// Prometheus metrics
var (
	// HTTP request counter
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP request duration histogram
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP error counter
	httpErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Active connections gauge
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		},
	)

	// Logs received counter
	logsReceivedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "logs_received_total",
			Help: "Total number of logs received",
		},
		[]string{"client_id", "module"},
	)
)

/**
 * InitMetrics initializes Prometheus metrics
 * @description
 * - Initializes all Prometheus metrics
 * - Registers metrics with Prometheus registry
 * - Sets default values for gauges
 * @throws
 * - Metrics registration errors
 */
func InitMetrics() {
	// Initialize active connections gauge
	activeConnections.Set(0)

	// Log metrics initialization
	logrus.Info("Prometheus metrics initialized")
}

/**
 * IncrementRequestCount increments the total request counter
 * @description
 * - Increments the global request counter
 * - Updates the active connections gauge
 * - Used by the request middleware
 */
func IncrementRequestCount() {
	// Increment utils counter
	utils.IncrementRequestCount()

	// Increment active connections
	activeConnections.Inc()
}

/**
 * DecrementActiveConnections decrements the active connections gauge
 * @description
 * - Decrements the active connections gauge
 * - Should be called when request processing completes
 */
func DecrementActiveConnections() {
	activeConnections.Dec()
}

/**
 * RecordHTTPRequest records HTTP request metrics
 * @param {string} method - HTTP method
 * @param {string} endpoint - Request endpoint
 * @param {int} statusCode - HTTP status code
 * @param {time.Duration} duration - Request duration
 * @description
 * - Records HTTP request count and duration
 * - Updates both total counter and histogram
 * - Formats status code as string for labels
 */
func RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration) {
	statusStr := strconv.Itoa(statusCode)

	// Increment request counter
	httpRequestsTotal.WithLabelValues(method, endpoint, statusStr).Inc()

	// Record request duration
	httpRequestDuration.WithLabelValues(method, endpoint, statusStr).Observe(duration.Seconds())

	// Record error if status code indicates error
	if statusCode >= 400 {
		httpErrorsTotal.WithLabelValues(method, endpoint, statusStr).Inc()
		utils.IncrementErrorCount()
	}
}

/**
 * RecordLogsReceived records logs received metrics
 * @param {string} clientID - Client identifier
 * @param {string} module - Module name
 * @description
 * - Records logs received count
 * - Updates the logs counter
 * - Used for logging analytics
 */
func RecordLogsReceived(clientID, module string) {
	logsReceivedTotal.WithLabelValues(clientID, module).Inc()
}
