package controllers

import (
	"costrict-keeper/internal/models"
	"costrict-keeper/services"
	"fmt"

	"github.com/gin-gonic/gin"
)

type ComponentController struct {
	component *services.ComponentManager
}

/**
 * Create new Component controller instance
 * @param {*services.ComponentManager} component - Component manager instance for managing components
 * @returns {*ComponentController} New Component controller instance
 * @description
 * - Initializes controller with component manager
 * - Used to manage API routes and handlers for component operations
 * @example
 * compManager := services.GetComponentManager()
 * controller := controllers.NewComponentController(compManager)
 */
func NewComponentController(component *services.ComponentManager) *ComponentController {
	return &ComponentController{
		component: component,
	}
}

/**
 * Register all component API routes to Gin router group
 * @param {*gin.RouterGroup} r - Gin router group instance
 * @description
 * - Registers routes for:
 *   - Component management (list/upgrade/delete)
 * @example
 * api := router.Group("/costrict/api/v1")
 * controller := NewComponentController(compManager)
 * controller.RegisterRoutes(api)
 */
func (c *ComponentController) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/costrict/api/v1")
	// 组件管理接口
	api.GET("/components", c.ListComponents)
	api.GET("/components/:name", c.GetComponentDetail)
	api.POST("/components/:name/upgrade", c.UpgradeComponent)
	api.DELETE("/components/:name", c.DeleteComponent)
}

// @Summary 获取组件列表
// @Description 获取所有已安装组件信息
// @Tags Components
// @Produce json
// @Success 200 {array} models.ComponentDetail
// @Router /costrict/api/v1/components [get]
func (c *ComponentController) ListComponents(g *gin.Context) {
	var components []models.ComponentDetail
	for _, ci := range c.component.GetComponents(true, true) {
		components = append(components, ci.GetDetail())
	}
	g.JSON(200, components)
}

// @Summary 升级组件
// @Description 升级指定组件到最新版本
// @Tags Components
// @Param name path string true "组件名称"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} models.ErrorResponse
// @Router /costrict/api/v1/components/{name}/upgrade [post]
func (c *ComponentController) UpgradeComponent(g *gin.Context) {
	name := g.Param("name")
	if err := c.component.UpgradeComponent(name); err != nil {
		if err == services.ErrComponentNotFound {
			g.JSON(404, &models.ErrorResponse{
				Code:  "component.not_found",
				Error: fmt.Sprintf("Component [%s] isn't exist", name),
			})
		} else {
			g.JSON(500, &models.ErrorResponse{
				Code:  "component.upgrade_failed",
				Error: err.Error(),
			})
		}
		return
	}
	g.JSON(200, gin.H{"status": "success"})
}

// @Summary 获取组件详情
// @Description 根据组件名称获取指定组件的详细信息
// @Tags Components
// @Param name path string true "组件名称"
// @Success 200 {object} models.ComponentDetail
// @Failure 404 {object} models.ErrorResponse
// @Router /costrict/api/v1/components/{name} [get]
func (c *ComponentController) GetComponentDetail(g *gin.Context) {
	name := g.Param("name")
	ci := c.component.GetComponent(name)
	if ci == nil {
		g.JSON(404, &models.ErrorResponse{
			Code:  "component.not_found",
			Error: fmt.Sprintf("Component [%s] isn't exist", name),
		})
		return
	}
	g.JSON(200, ci.GetDetail())
}

// @Summary 删除组件
// @Description 根据组件名删除指定组件
// @Tags Components
// @Param name path string true "组件名称"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} models.ErrorResponse
// @Router /costrict/api/v1/components/{name} [delete]
func (c *ComponentController) DeleteComponent(g *gin.Context) {
	_ = g.Param("name")

	// 注意：这里需要实现删除组件的逻辑
	// 目前先返回成功状态，实际项目中需要实现具体的删除逻辑
	g.JSON(404, &models.ErrorResponse{
		Code:  "component.not_implemented",
		Error: "component deletion not implemented yet",
	})
}
