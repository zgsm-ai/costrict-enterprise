package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"costrict-keeper/internal/logger"
	"costrict-keeper/internal/models"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_request_total",
			Help: "Total service requests",
		},
		[]string{"service"},
	)

	errorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_error_total",
			Help: "Total service error requests",
		},
		[]string{"service"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "service_request_duration_seconds",
			Help:    "Duration of service requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service"},
	)

	serviceHealthStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_health_status",
			Help: "Health status of services (1: healthy, 0: unhealthy)",
		},
		[]string{"service", "version"},
	)

	componentVersionInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "component_version_info",
			Help: "Version information of components",
		},
		[]string{"component", "version"},
	)

	serviceUpTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_uptime_seconds",
			Help: "Service uptime in seconds",
		},
		[]string{"service"},
	)

	// 本地计数器，用于快速获取总请求数
	totalRequests int64 = 0
	totalErrors   int64 = 0
)

func init() {
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(errorCount)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(serviceHealthStatus)
	prometheus.MustRegister(componentVersionInfo)
	prometheus.MustRegister(serviceUpTime)
}

/**
 * Collect metrics from all components
 * @returns {error} Returns error if collection fails, nil on success
 * @description
 * - Creates service manager instance to access component information
 * - Collects component health status and version information
 * - Collects service metrics including uptime and request counts
 * - Updates Prometheus gauge metrics for each component
 * @throws
 * - Service manager creation errors
 * - Component retrieval errors
 * - Health check errors
 */
func collectMetricsFromComponents() error {
	// Create service manager to access component information
	sm := GetServiceManager()

	// Collect metrics for each service
	services := sm.GetInstances(true)
	for _, service := range services {
		// Set component health status (1: healthy, 0: unhealthy)
		svc := service.GetDetail()
		healthStatus := 0.0
		if svc.Component != nil && svc.Component.Installed {
			healthStatus = 1.0
		}
		cpn := svc.Component
		if cpn != nil {
			serviceHealthStatus.WithLabelValues(svc.Name, cpn.Local.Version).Set(healthStatus)

			// Set cpn version info (using value 1 as placeholder since version is already in label)
			componentVersionInfo.WithLabelValues(svc.Name, cpn.Local.Version).Set(1.0)

			logger.Debugf("Collected metrics for component %s, version: %s, installed: %v",
				svc.Name, cpn.Local.Version, cpn.Installed)
		}

		// Check if svc is healthy
		healthy := service.GetHealthy()
		healthValue := 0.0
		if healthy == models.Healthy {
			healthValue = 1.0
		}
		serviceHealthStatus.WithLabelValues(svc.Name, "unknown").Set(healthValue)

		// If svc has metrics endpoint, try to collect additional metrics
		if svc.Spec.Metrics != "" && svc.Port > 0 {
			if err := collectServiceMetrics(svc.Spec); err != nil {
				logger.Warnf("Failed to collect metrics from service %s: %v", svc.Name, err)
			}
		}

		logger.Debugf("Collected metrics for service %s, healthy: %v", svc.Name, healthy)
	}

	return nil
}

/**
 * Collect additional metrics from a specific service
 * @param {models.ServiceSpecification} service - Service specification
 * @returns {error} Returns error if collection fails, nil on success
 * @description
 * - Constructs service metrics endpoint URL
 * - Makes HTTP request to service metrics endpoint
 * - Processes and records service-specific metrics
 * @throws
 * - HTTP request errors
 * - Response parsing errors
 */
func collectServiceMetrics(service models.ServiceSpecification) error {
	// Construct metrics URL
	url := fmt.Sprintf("http://localhost:%d%s", service.Port, service.Metrics)

	// Create HTTP client with timeout
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}

	// Ensure connection pool is cleaned up
	defer tr.CloseIdleConnections()

	// Make HTTP request to service metrics endpoint
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get metrics from service %s: %v", service.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("service %s returned non-200 status code: %d", service.Name, resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body from service %s: %v", service.Name, err)
	}

	// For now, just log the metrics content
	// In a real implementation, you would parse the metrics and update Prometheus counters
	logger.Debugf("Metrics from service %s: %s", service.Name, string(body))

	return nil
}

/**
 * Push collected metrics to Prometheus Pushgateway
 * @param {string} addr - Pushgateway address
 * @returns {error} Returns error if push fails, nil on success
 * @description
 * - Creates pusher instance with specified gateway address
 * - Pushes all registered Prometheus metrics to gateway
 * - Handles push errors and logging
 * @throws
 * - Pushgateway connection errors
 * - Push operation errors
 */
func pushMetricsToGateway(addr string) error {
	if addr == "" {
		return fmt.Errorf("pushgateway address is empty")
	}

	// Create a pusher to push metrics to the pushgateway
	pusher := push.New(addr, "costrict")

	// Add default metrics
	pusher.Collector(requestCount)
	pusher.Collector(requestDuration)
	pusher.Collector(serviceHealthStatus)
	pusher.Collector(componentVersionInfo)
	pusher.Collector(serviceUpTime)

	// Push metrics to gateway
	if err := pusher.Add(); err != nil {
		logger.Errorf("Failed to push metrics to pushgateway: %v", err)
		return err
	}

	logger.Infof("Successfully pushed metrics to pushgateway: %s", addr)
	return nil
}

/**
 * Collect and push metrics periodically
 * @param {string} pushGatewayAddr - Pushgateway address
 * @returns {error} Returns error if operation fails, nil on success
 * @description
 * - Initializes metrics collection and push process
 * - Sets up periodic ticker for regular metric collection
 * - Handles context cancellation for graceful shutdown
 * - Executes initial collection and push immediately
 * @throws
 * - Initial collection errors
 * - Initial push errors
 */
func CollectAndPushMetrics(pushGatewayAddr string) error {
	fmt.Println("启动指标采集服务(无服务器模式)，Pushgateway地址:", pushGatewayAddr)

	ctx := context.Background()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 执行一次指标采集和推送
	if err := collectMetricsFromComponents(); err != nil {
		fmt.Printf("指标采集失败: %v\n", err)
		return err
	}

	if err := pushMetricsToGateway(pushGatewayAddr); err != nil {
		fmt.Printf("指标推送失败: %v\n", err)
		return err
	}

	select {
	case <-ticker.C:
		return nil
	case <-ctx.Done():
		return nil
	}
}

/**
 * Increment request counter for a specific service
 * @param {string} serviceName - Name of the service
 * @description
 * - Increments the request counter for the specified service
 * - Used by API handlers to track request counts
 */
func IncrementRequestCount(serviceName string) {
	requestCount.WithLabelValues(serviceName).Inc()
	IncrementTotalRequestCount()
}

/**
 * Record request duration for a specific service
 * @param {string} serviceName - Name of the service
 * @param {float64} duration - Request duration in seconds
 * @description
 * - Records the duration of a request for the specified service
 * - Used by API handlers to track request latency
 */
func RecordRequestDuration(serviceName string, duration float64) {
	requestDuration.WithLabelValues(serviceName).Observe(duration)
}

/**
 * Update service uptime metric
 * @param {string} serviceName - Name of the service
 * @param {float64} uptime - Service uptime in seconds
 * @description
 * - Updates the uptime metric for the specified service
 * - Used by service manager to track service availability
 */
func UpdateServiceUptime(serviceName string, uptime float64) {
	serviceUpTime.WithLabelValues(serviceName).Set(uptime)
}

/**
 * Increment error counter for a specific service
 * @param {string} serviceName - Name of the service
 * @description
 * - Increments the error counter for the specified service
 * - Used by API handlers to track error request counts
 */
func IncrementErrorCount(serviceName string) {
	errorCount.WithLabelValues(serviceName).Inc()
	totalErrors++
}

/**
 * Get total request count
 * @returns {int64} Returns total request count
 * @description
 * - Returns the total number of requests received
 * - Used by health check endpoint
 */
func GetTotalRequestCount() int64 {
	return totalRequests
}

/**
 * Get total error count
 * @returns {int64} Returns total error count
 * @description
 * - Returns the total number of error requests received
 * - Used by health check endpoint
 */
func GetTotalErrorCount() int64 {
	return totalErrors
}

/**
 * Increment total request count
 * @description
 * - Increments the total request counter
 * - Used by middleware to track overall request count
 */
func IncrementTotalRequestCount() {
	totalRequests++
}
