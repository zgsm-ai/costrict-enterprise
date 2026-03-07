package types

import (
	"time"
)

// CodebaseSyncMessage 表示从 Redis 消息队列接收的代码库同步消息
type CodebaseSyncMessage struct {
	SyncID             int32     `json:"syncId"`             // 同步操作ID
	CodebaseID         int32     `json:"codebaseId"`         // 代码库ID
	CodebasePath       string    `json:"codebasePath"`       // 代码库路径
	CodebaseName       string    `json:"codebaseName"`       // 代码库名字
	SyncTime           time.Time `json:"syncTime"`           // 同步结束时间
	IsEmbedTaskSuccess bool      `json:"isEmbedTaskSuccess"` // 嵌入任务是否成功
	IsGraphTaskSuccess bool      `json:"isGraphTaskSuccess"` // 图任务是否成功
	FailedTimes        int       `json:"failedTimes"`
}

const (
	TaskTypeCodegraph = "codegraph"
	TaskTypeEmbedding = "embedding"
)

// IndexMessage 索引任务消息
type IndexMessage struct {
	CodebaseId int32  `json:"codebase_id"`
	Path       string `json:"path"`
}
