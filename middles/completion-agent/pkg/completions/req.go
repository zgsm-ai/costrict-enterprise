package completions

// 补全请求结构
type CompletionRequest struct {
	Model        string                 `json:"model,omitempty"`
	LanguageID   string                 `json:"language_id,omitempty"`
	ClientID     string                 `json:"client_id,omitempty"`
	CompletionID string                 `json:"completion_id,omitempty"`
	Temperature  float64                `json:"temperature,omitempty"`
	TriggerMode  string                 `json:"trigger_mode,omitempty"`
	ParentID     string                 `json:"parent_id,omitempty"`
	Stop         []string               `json:"stop,omitempty"`
	Verbose      bool                   `json:"verbose,omitempty"`
	Extra        map[string]interface{} `json:"extra,omitempty"`
	Prompts      *PromptOptions         `json:"prompt_options,omitempty"`
	HideScores   *HiddenScoreOptions    `json:"calculate_hide_score,omitempty"`
}

type Snippet struct {
	Type     string `json:"type"`
	Content  string `json:"content"`
	FilePath string `json:"filepath,omitempty"`
	CopiedAt string `json:"copiedAt,omitempty"`
}

// 提示词选项
type PromptOptions struct {
	Prefix                string    `json:"prefix,omitempty"`                  // The code snippet before the cursor
	Suffix                string    `json:"suffix,omitempty"`                  // The code snippet after the cursor
	CodeContext           string    `json:"code_context,omitempty"`            // Obsolete
	ProjectPath           string    `json:"project_path,omitempty"`            // Root Path Snippets
	FileProjectPath       string    `json:"file_project_path,omitempty"`       //
	ImportContent         string    `json:"import_content,omitempty"`          // Import Definition Snippets
	RecentlyEditedRanges  []Snippet `json:"recently_edited_ranges,omitempty"`  // Recently Edited Range Snippets
	RecentlyVisitedRanges []Snippet `json:"recently_visited_ranges,omitempty"` // Recently Visited Range Snippets
	ClipboardContent      []Snippet `json:"clipboard_content,omitempty"`       // Clipboard Snippets
	RecentlyOpenedFiles   []Snippet `json:"recently_opened_files,omitempty"`   // Recently Opened Files Snippets
	StaticContext         []Snippet `json:"static_context,omitempty"`          // Static Snippets
}

// 计算隐藏分数配置
type HiddenScoreOptions struct {
	IsWhitespaceAfterCursor bool  `json:"is_whitespace_after_cursor"` //光标之后该行是否没有内容(空白除外)
	DocumentLength          int   `json:"document_length"`            //文档长度
	PromptEndPos            int   `json:"prompt_end_pos"`             //光标在文档中的偏移
	PreviousLabel           int   `json:"previous_label"`             //上个请求是否被接受
	PreviousLabelTimestamp  int64 `json:"previous_label_timestamp"`   //上个请求被接受的时间戳
}
