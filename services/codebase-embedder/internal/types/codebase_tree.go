package types

// TreeNode 目录树节点
type TreeNode struct {
	Name     string      `json:"name"`               // 节点名称
	Path     string      `json:"path"`               // 完整路径
	Type     string      `json:"type"`               // 节点类型: file/directory
	Children []*TreeNode `json:"children,omitempty"` // 子节点，仅目录节点有效
}

// CodebaseTreeRequest 目录树查询请求
type CodebaseTreeRequest struct {
	ClientId     string `json:"clientId"`     // 客户端唯一标识
	CodebasePath string `json:"codebasePath"` // 项目绝对路径
	CodebaseName string `json:"codebaseName"` // 项目名称
	MaxDepth     *int   `json:"maxDepth"`     // 目录树最大深度，可选
	IncludeFiles *bool  `json:"includeFiles"` // 是否包含文件节点，可选
}

// CodebaseTreeResponse 目录树查询响应
type CodebaseTreeResponse struct {
	Code    int       `json:"code"`    // 响应码
	Message string    `json:"message"` // 响应消息
	Success bool      `json:"success"` // 是否成功
	Data    *TreeNode `json:"data"`    // 目录树数据
}
