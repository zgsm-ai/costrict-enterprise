package completions

// 补全请求结构
type CompletionRequest struct {
	Model           string                 `json:"model,omitempty"`
	Prompt          string                 `json:"prompt"`                      //废弃
	ProjectPath     string                 `json:"project_path,omitempty"`      //废弃
	FileProjectPath string                 `json:"file_project_path,omitempty"` //废弃
	ImportContent   string                 `json:"import_content,omitempty"`    //废弃
	BetaMode        bool                   `json:"beta_mode,omitempty"`         //废弃
	LanguageID      string                 `json:"language_id,omitempty"`
	ClientID        string                 `json:"client_id,omitempty"`
	CompletionID    string                 `json:"completion_id,omitempty"`
	Temperature     float64                `json:"temperature,omitempty"`
	TriggerMode     string                 `json:"trigger_mode,omitempty"`
	ParentID        string                 `json:"parent_id,omitempty"`
	Stop            []string               `json:"stop,omitempty"`
	Verbose         bool                   `json:"verbose,omitempty"`
	Extra           map[string]interface{} `json:"extra,omitempty"`
	Prompts         *PromptOptions         `json:"prompt_options,omitempty"`
	HideScores      *HiddenScoreOptions    `json:"calculate_hide_score,omitempty"`
}

// 提示词选项
type PromptOptions struct {
	Prefix          string `json:"prefix,omitempty"`
	Suffix          string `json:"suffix,omitempty"`
	CodeContext     string `json:"code_context,omitempty"`
	ProjectPath     string `json:"project_path,omitempty"`
	FileProjectPath string `json:"file_project_path,omitempty"`
	ImportContent   string `json:"import_content,omitempty"`
}

// 计算隐藏分数配置
type HiddenScoreOptions struct {
	IsWhitespaceAfterCursor bool   `json:"is_whitespace_after_cursor"` //光标之后该行是否没有内容(空白除外)
	Prefix                  string `json:"prefix,omitempty"`           //光标前的所有内容(废弃)
	DocumentLength          int    `json:"document_length"`            //文档长度
	PromptEndPos            int    `json:"prompt_end_pos"`             //光标在文档中的偏移
	PreviousLabel           int    `json:"previous_label"`             //上个请求是否被接受
	PreviousLabelTimestamp  int64  `json:"previous_label_timestamp"`   //上个请求被接受的时间戳
}
