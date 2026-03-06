package controllers

import (
	"costrict-keeper/internal/config"
	"costrict-keeper/internal/models"
	"costrict-keeper/services"

	"github.com/gin-gonic/gin"
)

type APIController struct {
	server *services.Server
}

/**
 * Create new API controller instance
 * @param {*services.ServiceManager} svc - Service manager instance for managing services
 * @returns {*APIController} New API controller instance
 * @description
 * - Initializes controller with service manager
 * - Used to manage API routes and handlers for service operations
 * @example
 * svcManager := services.GetServiceManager()
 * controller := controllers.NewAPIController(svcManager)
 */
func NewAPIController(server *services.Server) *APIController {
	return &APIController{
		server: server,
	}
}

/**
 * Register all API routes to Gin engine
 * @param {*gin.Engine} r - Gin router instance
 * @description
 * - Creates /api route group
 * - Registers routes for:
 *   - Service management (list/restart)
 *   - Component management (list/upgrade)
 *   - Endpoint listing
 * @example
 * router := gin.Default()
 * controller := NewAPIController(svcManager)
 * controller.RegisterRoutes(router)
 */
func (a *APIController) RegisterRoutes(r *gin.Engine) {
	r.GET("/healthz", a.Healthz)
	r.GET("/costrict/api/v1/state", a.GetState)
	r.POST("/costrict/api/v1/reload", a.ReloadConfig)
	r.POST("/costrict/api/v1/check", a.Check)
}

// @Summary 获取服务器状态
// @Description 获取服务器状态信息，包括系统规格、认证配置、软件配置和云配置，端口分配信息，等
// @Tags Config
// @Accept json
// @Produce json
// @Success 200 {object} models.ServerState "服务器状态"
// @Router /costrict/api/v1/state [get]
func (a *APIController) GetState(c *gin.Context) {
	c.JSON(200, a.server.GetState())
}

// @Summary 重新加载配置
// @Description 重新加载应用配置文件
// @Tags Config
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /costrict/api/v1/reload [post]
func (a *APIController) ReloadConfig(c *gin.Context) {
	// 调用配置重新加载方法
	if err := config.ReloadConfig(false); err != nil {
		c.JSON(500, &models.ErrorResponse{
			Code:  "config.reload_failed",
			Error: "Failed to reload configuration: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}

// @Summary 执行系统检查
// @Description 立即执行各项检查，包括服务健康状态、进程状态、隧道状态、组件更新状态和半夜鸡叫自动升级检查机制
// @Description 返回详细的检查结果，包括各项服务的运行状态、进程信息、隧道连接状态、组件版本信息以及系统总体健康状态，但不包含配置信息
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} models.CheckResponse "检查成功，返回详细的系统状态信息（不包含配置信息）"
// @Success 200 {object} models.CheckResponse "示例响应：{\n  \"timestamp\": \"2024-01-01T10:00:00Z\",\n  \"services\": [{\n    \"name\": \"costrict\",\n    \"status\": \"running\",\n    \"pid\": 1234,\n    \"port\": 8080,\n    \"startTime\": \"2024-01-01T09:00:00Z\",\n    \"healthy\": true\n  }],\n  \"processes\": [],\n  \"tunnels\": [{\n    \"name\": \"myapp\",\n    \"localPort\": 8080,\n    \"mappingPort\": 30001,\n    \"status\": \"running\",\n    \"pid\": 1235,\n    \"createdTime\": \"2024-01-01T09:00:00Z\"\n  }],\n  \"components\": [{\n    \"name\": \"costrict\",\n    \"localVersion\": \"1.0.0\",\n    \"remoteVersion\": \"1.1.0\",\n    \"installed\": true,\n    \"needUpgrade\": true\n  }],\n  \"midnightRooster\": {\n    \"status\": \"active\",\n    \"nextCheckTime\": \"2024-01-02T03:30:00Z\",\n    \"lastCheckTime\": \"2024-01-01T03:30:00Z\",\n    \"componentsCount\": 5,\n    \"upgradesNeeded\": 2\n  },\n  \"overallStatus\": \"warning\",\n  \"totalChecks\": 4,\n  \"passedChecks\": 3,\n  \"failedChecks\": 1\n}"
// @Router /costrict/api/v1/check [post]
func (a *APIController) Check(c *gin.Context) {
	// 调用server的Check方法执行系统检查
	response := a.server.Check()
	c.JSON(200, response)
}

// @Summary 业务就绪探针
// @Description 检查服务是否已经做好准备，返回服务版本、启动时间、健康状态和关键指标统计结果
// @Tags System
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Router /healthz [get]
func (a *APIController) Healthz(c *gin.Context) {
	// 调用server的GetHealthz方法获取健康检查响应
	response := a.server.GetHealthz()
	c.JSON(200, response)
}
