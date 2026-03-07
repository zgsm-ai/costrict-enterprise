package types

// FileStatusRequest 文件状态查询请求
type FileStatusRequest struct {
	ClientId     string `json:"clientId"`                       // 客户ID，machineID
	CodebasePath string `json:"codebasePath"`                   // 项目绝对路径（按照操作系统格式）
	CodebaseName string `json:"codebaseName"`                   // 项目名称
	ChunkNumber  int    `json:"chunkNumber,optional,default=0"` // 当前分片
	TotalChunks  int    `json:"totalChunks,optional,default=1"` // 分片总数，当代码过大时候采用分片上传（默认为1）
	SyncId       string `json:"syncId"`                         // 上传接口的RequestId
}

// FileStatusResponseData 文件状态查询响应数据
type FileStatusResponseData struct {
	Process       string           `json:"process"`       // 整体提取状态（如：pending/processing/complete/failed）
	TotalProgress int              `json:"totalProgress"` // 当前分片整体提取进度（百分比，0-100）
	FileList      []FileStatusItem `json:"fileList"`      // 文件列表
}

// FileStatusItem 单个文件状态项
type FileStatusItem struct {
	Path    string `json:"path"`    // 文件路径
	Status  string `json:"status"`  // 文件状态（如：pending/processing/complete/failed）
	Operate string `json:"operate"` // 文件操作类型（如：add/modify/delete）
}
