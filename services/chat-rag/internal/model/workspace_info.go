package model

import (
	"net/http"
	"strings"
)

// WorkspaceInfo is a snapshot of request attribution headers shared by logs and
// metrics reports. All fields are serialized, including missing values.
type WorkspaceInfo struct {
	SessionID       string `json:"session_id"`
	ZgsmTaskID      string `json:"zgsm_task_id"`
	TaskID          string `json:"task_id"`
	GitURL          string `json:"git_url"`
	GitCommit       string `json:"git_commit"`
	ProjectPath     string `json:"project_path"`
	ProjectID       string `json:"project_id"`
	WorkspaceID     string `json:"workspace_id"`
	WorkflowTaskID  string `json:"workflow_task_id"`
	IssueID         string `json:"issue_id"`
	NodeRunID       string `json:"node_run_id"`
	RequestID       string `json:"request_id"`
	KnownClient     string `json:"known_client"`
	ClientUserAgent string `json:"client_user_agent"`
	UserAgent       string `json:"user_agent"`
	AgentType       string `json:"agent_type"`
	CostrictVersion string `json:"costrict_version"`
}

// NewWorkspaceInfo reads only incoming headers, without deriving missing values
// from legacy identity fields. A session ID takes precedence over a legacy task ID.
func NewWorkspaceInfo(headers http.Header) WorkspaceInfo {
	info := WorkspaceInfo{
		SessionID:       strings.TrimSpace(headers.Get("X-Session-Id")),
		ZgsmTaskID:      strings.TrimSpace(headers.Get("zgsm-task-id")),
		GitURL:          strings.TrimSpace(headers.Get("X-Git-Url")),
		GitCommit:       strings.TrimSpace(headers.Get("X-Git-Commit")),
		ProjectPath:     strings.TrimSpace(headers.Get("zgsm-project-path")),
		ProjectID:       strings.TrimSpace(headers.Get("X-Project-Id")),
		WorkspaceID:     strings.TrimSpace(headers.Get("X-Workspace-Id")),
		WorkflowTaskID:  strings.TrimSpace(headers.Get("X-Workflow-Task-Id")),
		IssueID:         strings.TrimSpace(headers.Get("X-Issue-Id")),
		NodeRunID:       strings.TrimSpace(headers.Get("X-Node-Run-Id")),
		RequestID:       strings.TrimSpace(headers.Get("X-Request-Id")),
		KnownClient:     strings.TrimSpace(headers.Get("X-Known-Client")),
		ClientUserAgent: strings.TrimSpace(headers.Get("Client-UserAgent")),
		UserAgent:       strings.TrimSpace(headers.Get("User-Agent")),
		AgentType:       strings.TrimSpace(headers.Get("agent-type")),
		CostrictVersion: strings.TrimSpace(headers.Get("X-Costrict-Version")),
	}
	info.TaskID = info.SessionID
	if info.TaskID == "" {
		info.TaskID = info.ZgsmTaskID
	}
	return info
}
