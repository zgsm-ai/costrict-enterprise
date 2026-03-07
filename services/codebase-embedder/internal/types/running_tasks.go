package types

import "time"

// RunningTaskInfo 运行中任务信息
type RunningTaskInfo struct {
	TaskId                  string           `json:"taskId"`                            // 任务ID
	ClientId                string           `json:"clientId"`                          // 客户端ID
	Status                  string           `json:"status"`                            // 任务状态
	Process                 string           `json:"process"`                           // 处理进程
	TotalProgress           int              `json:"totalProgress"`                     // 总进度
	StartTime               time.Time        `json:"startTime"`                         // 开始时间
	LastUpdateTime          time.Time        `json:"lastUpdateTime"`                    // 最后更新时间
	EstimatedCompletionTime *time.Time       `json:"estimatedCompletionTime,omitempty"` // 预计完成时间
	FileList                []FileStatusItem `json:"fileList"`                          // 文件列表
}

// RunningTasksResponse 运行中任务查询响应
type RunningTasksResponse struct {
	Code    int               `json:"code"`    // 响应码
	Message string            `json:"message"` // 响应消息
	Success bool              `json:"success"` // 是否成功
	Data    *RunningTasksData `json:"data"`    // 任务数据
}

// RunningTasksData 运行中任务数据
type RunningTasksData struct {
	TotalTasks int               `json:"totalTasks"` // 任务总数
	Tasks      []RunningTaskInfo `json:"tasks"`      // 任务列表
}
