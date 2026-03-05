package router

import (
	"github.com/zgsm-ai/client-manager/controllers"
	_ "github.com/zgsm-ai/client-manager/docs"
	"github.com/zgsm-ai/client-manager/internal"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all routes for the application
/**
 * Setup all routes for the application
 * @param {*gin.Engine} r - Gin engine
 * @param {*controllers.LogController} logController - Log controller
 * @param {*logrus.Logger} logger - Application logger
 * @description
 * - Adds CORS middleware
 * - Adds Prometheus middleware
 * - Adds request ID middleware
 * - Sets up health check endpoints
 * - Sets up metrics endpoint
 * - Sets up Swagger documentation endpoint
 * - Sets up API routes
 */
func SetupRoutes(r *gin.Engine, logController *controllers.LogController, logger *logrus.Logger) {
	// Add CORS middleware
	r.Use(internal.CORSMiddleware())

	// Add Prometheus middleware
	r.Use(internal.PrometheusMiddleware())

	// Add request ID middleware
	r.Use(internal.RequestIDMiddleware())

	// Health check endpoints
	setupHealthCheckRoutes(r, logger)

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup API routes
	setupAPIRoutes(r, logController)
}

// setupHealthCheckRoutes configures health check routes
/**
 * Setup health check routes
 * @param {*gin.Engine} r - Gin engine
 * @param {*logrus.Logger} logger - Application logger
 * @description
 * - Sets up /healthz endpoint
 * - Sets up /live endpoint
 * - Sets up /ready endpoint
 */
func setupHealthCheckRoutes(r *gin.Engine, logger *logrus.Logger) {
	healthController := controllers.NewHealthController(logger)

	r.GET("/healthz", healthController.GetHealth)
	r.GET("/live", healthController.LiveHandler)
	r.GET("/ready", healthController.ReadyHandler)
}

// setupAPIRoutes configures API routes for the application
/**
 * Setup API routes for the application
 * @param {*gin.Engine} r - Gin engine
 * @param {*controllers.LogController} logController - Log controller
 * @description
 * - Sets up configuration API routes
 * - Sets up feedback API routes
 * - Sets up log API routes
 */
func setupAPIRoutes(r *gin.Engine, logController *controllers.LogController) {
	// Setup API routes
	api := r.Group("/client-manager/api/v1")
	{
		// Log routes
		logs := api.Group("/logs")
		{
			logs.POST("", logController.PostLog)
			logs.GET("", logController.ListLogs)
			logs.GET("/:client_id/:file_name", logController.GetLogs)
		}
	}
}
