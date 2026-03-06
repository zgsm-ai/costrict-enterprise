package models

import (
	"time"
)

// CheckResponse 检查API响应结构
// @Description 系统检查API响应数据结构
type CheckResponse struct {
	Timestamp     time.Time         `json:"timestamp" example:"2024-01-01T10:00:00Z" description:"检查时间戳"`
	Services      []ServiceDetail   `json:"services" description:"服务检查结果列表"`
	Components    []ComponentDetail `json:"components" description:"组件检查结果列表"`
	OverallStatus string            `json:"overallStatus" description:"总体状态"`
	TotalChecks   int               `json:"totalChecks" description:"总检查项数"`
	PassedChecks  int               `json:"passedChecks" description:"通过检查项数"`
	FailedChecks  int               `json:"failedChecks" description:"失败检查项数"`
}
