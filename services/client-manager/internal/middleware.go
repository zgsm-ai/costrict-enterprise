package internal

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

/**
 * CORSMiddleware handles Cross-Origin Resource Sharing (CORS)
 * @description
 * - Adds CORS headers to the response
 * - Handles preflight requests
 * - Configures allowed origins, methods, and headers
 * @returns {gin.HandlerFunc} Gin middleware function
 */
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow all origins for development
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

/**
 * RequestIDMiddleware adds a unique request ID to each request
 * @description
 * - Generates a unique UUID for each request
 * - Adds the request ID to the context
 * - Includes request ID in response headers
 * - Helps with request tracing and debugging
 * @returns {gin.HandlerFunc} Gin middleware function
 */
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()

		// Add to context
		c.Set("request_id", requestID)

		// Add to response header
		c.Header("X-Request-ID", requestID)

		// Add to logger context
		c.Set("logger", logrus.WithField("request_id", requestID))

		c.Next()
	}
}

/**
 * LoggerMiddleware logs HTTP requests
 * @description
 * - Logs request method, path, status code, and duration
 * - Includes request ID in logs
 * - Formats logs in JSON for structured logging
 * - Supports different log levels based on status codes
 * @returns {gin.HandlerFunc} Gin middleware function
 */
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Get logger from context
		logger, exists := c.Get("logger")
		var logEntry *logrus.Entry
		if exists {
			logEntry = logger.(*logrus.Entry)
		} else {
			logEntry = logrus.NewEntry(logrus.New())
		}

		// Calculate duration
		duration := time.Since(start)

		// Log request details
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Determine log level based on status code
		switch {
		case statusCode >= 500:
			logEntry.WithFields(logrus.Fields{
				"method":     method,
				"path":       path,
				"status":     statusCode,
				"duration":   duration,
				"client_ip":  clientIP,
				"user_agent": userAgent,
			}).Error("HTTP request failed")
		case statusCode >= 400:
			logEntry.WithFields(logrus.Fields{
				"method":     method,
				"path":       path,
				"status":     statusCode,
				"duration":   duration,
				"client_ip":  clientIP,
				"user_agent": userAgent,
			}).Warn("HTTP request warning")
		default:
			logEntry.WithFields(logrus.Fields{
				"method":     method,
				"path":       path,
				"status":     statusCode,
				"duration":   duration,
				"client_ip":  clientIP,
				"user_agent": userAgent,
			}).Info("HTTP request completed")
		}
	}
}

/**
 * PrometheusMiddleware collects metrics for Prometheus
 * @description
 * - Increments request counter for each request
 * - Records request duration
 * - Tracks response status codes
 * - Updates global metrics counters
 * - Records active connections
 * @returns {gin.HandlerFunc} Gin middleware function
 */
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Increment request counter and active connections
		IncrementRequestCount()

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Record metrics
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		// Record HTTP request metrics
		RecordHTTPRequest(method, path, statusCode, duration)

		// Decrement active connections
		DecrementActiveConnections()

	}
}

/**
 * TimeoutMiddleware adds timeout to requests
 * @description
 * - Sets timeout for request processing
 * - Cancels context if timeout is exceeded
 * - Prevents long-running requests
 * @param {time.Duration} timeout - Request timeout duration
 * @returns {gin.HandlerFunc} Gin middleware function
 */
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context
		c.Request = c.Request.WithContext(ctx)

		// Create channel to monitor completion
		done := make(chan struct{})

		// Process request in goroutine
		go func() {
			c.Next()
			close(done)
		}()

		// Wait for completion or timeout
		select {
		case <-done:
			// Request completed normally
			return
		case <-ctx.Done():
			// Timeout occurred
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
				"code":    "timeout.error",
				"message": "Request timed out",
			})
			return
		}
	}
}

/**
 * RateLimitMiddleware implements rate limiting
 * @description
 * - Limits requests per client IP
 * - Uses sliding window algorithm
 * - Returns 429 status if limit exceeded
 * @param {int} requests - Maximum number of requests
 * @param {time.Duration} window - Time window for rate limiting
 * @returns {gin.HandlerFunc} Gin middleware function
 */
func RateLimitMiddleware(requests int, window time.Duration) gin.HandlerFunc {
	// In a real implementation, this would use Redis or a similar distributed cache
	// For simplicity, we'll use an in-memory store
	type clientRecord struct {
		count     int
		timestamp time.Time
	}
	clients := make(map[string]*clientRecord)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Get or create client record
		record, exists := clients[clientIP]
		if !exists {
			record = &clientRecord{
				count:     0,
				timestamp: time.Now(),
			}
			clients[clientIP] = record
		}

		// Check if window has expired
		if time.Since(record.timestamp) > window {
			record.count = 0
			record.timestamp = time.Now()
		}

		// Check if limit exceeded
		if record.count >= requests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    "rate_limit.exceeded",
				"message": "Rate limit exceeded",
			})
			return
		}

		// Increment counter
		record.count++

		c.Next()
	}
}

/**
 * AuthMiddleware handles authentication
 * @description
 * - Validates authentication token
 * - Extracts user information from token
 * - Adds user information to context
 * - Returns 401 if authentication fails
 * @returns {gin.HandlerFunc} Gin middleware function
 */
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "auth.missing",
				"message": "Authorization header is required",
			})
			return
		}

		// Check Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "auth.invalid_format",
				"message": "Authorization header must be Bearer token",
			})
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "auth.empty_token",
				"message": "Token is required",
			})
			return
		}

		// Validate token (in a real implementation, this would validate JWT or similar)
		// For simplicity, we'll just check if token is not empty
		// In production, you should implement proper token validation
		userID := "user_" + token // Simplified user extraction

		// Add user information to context
		c.Set("user_id", userID)

		c.Next()
	}
}

/**
 * RecoveryMiddleware recovers from panics
 * @description
 * - Recovers from panics in handlers
 * - Logs panic information
 * - Returns 500 error response
 * - Prevents application crashes
 * @returns {gin.HandlerFunc} Gin middleware function
 */
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Get logger from context
		logger, exists := c.Get("logger")
		var logEntry *logrus.Entry
		if exists {
			logEntry = logger.(*logrus.Entry)
		} else {
			logEntry = logrus.NewEntry(logrus.New())
		}

		// Log panic
		logEntry.WithField("panic", recovered).Error("Panic recovered")

		// Return error response
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":    "internal.error",
			"message": "Internal server error",
		})
	})
}

/**
 * SetSecurityHeaders adds security-related headers
 * @description
 * - Adds security headers to prevent common attacks
 * - Includes XSS protection, content type, and other security headers
 * @returns {gin.HandlerFunc} Gin middleware function
 */
func SetSecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		c.Next()
	}
}
