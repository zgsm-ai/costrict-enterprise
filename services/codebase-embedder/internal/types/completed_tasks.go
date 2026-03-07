package types

import "time"

// CompletedTaskInfo 已完成任务信息
type CompletedTaskInfo struct {
	TaskId        string           `json:"taskId"`        // 任务ID
	ClientId      string           `json:"clientId"`      // 客户端ID
	Status        string           `json:"status"`        // 任务状态
	Process       string           `json:"process"`       // 处理进程
	TotalProgress int              `json:"totalProgress"` // 总进度
	CompletedTime time.Time        `json:"completedTime"` // 完成时间
	FileCount     int              `json:"fileCount"`     // 文件总数
	SuccessCount  int              `json:"successCount"`  // 成功文件数
	FailedCount   int              `json:"failedCount"`   // 失败文件数
	SuccessRate   float64          `json:"successRate"`   // 成功率
	FileList      []FileStatusItem `json:"fileList"`      // 文件列表
}

// CompletedTasksResponse 已完成任务查询响应
type CompletedTasksResponse struct {
	Code    int                 `json:"code"`    // 响应码
	Message string              `json:"message"` // 响应消息
	Success bool                `json:"success"` // 是否成功
	Data    *CompletedTasksData `json:"data"`    // 任务数据
}

// CompletedTasksData 已完成任务数据
type CompletedTasksData struct {
	TotalTasks int                 `json:"totalTasks"` // 任务总数
	Tasks      []CompletedTaskInfo `json:"tasks"`      // 任务列表
}

// FailedTasksResponse 失败任务查询响应
type FailedTasksResponse struct {
	Code    int              `json:"code"`    // 响应码
	Message string           `json:"message"` // 响应消息
	Success bool             `json:"success"` // 是否成功
	Data    *FailedTasksData `json:"data"`    // 任务数据
}

// FailedTasksData 失败任务数据
type FailedTasksData struct {
	TotalTasks int                 `json:"totalTasks"` // 任务总数
	Tasks      []CompletedTaskInfo `json:"tasks"`      // 任务列表
}
