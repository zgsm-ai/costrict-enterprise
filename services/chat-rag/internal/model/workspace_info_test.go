package model

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestWorkspaceInfoHeadersAndLogJSON(t *testing.T) {
	headers := make(http.Header)
	inputs := map[string]string{
		"x-session-id": "session", "zgsm-task-id": "legacy",
		"X-Git-Url": "https://example.com/team/repo.git", "X-Git-Commit": "abc123",
		"zgsm-project-path": "%2Fproject", "X-Project-Id": "project", "X-Workspace-Id": "workspace",
		"X-Workflow-Task-Id": "workflow-task", "X-Issue-Id": "issue", "X-Node-Run-Id": "node",
		"X-Request-Id": "request", "X-Known-Client": "csc", "Client-UserAgent": "csc/1",
		"User-Agent": "bridge/1", "agent-type": "cs-bridge-gateway", "X-Costrict-Version": "version",
	}
	for key, value := range inputs {
		headers.Set(strings.ToLower(key), " \t"+value+" \t")
	}
	expected := map[string]string{
		"session_id": "session", "zgsm_task_id": "legacy", "task_id": "session",
		"git_url": "https://example.com/team/repo.git", "git_commit": "abc123", "project_path": "%2Fproject",
		"project_id": "project", "workspace_id": "workspace", "workflow_task_id": "workflow-task",
		"issue_id": "issue", "node_run_id": "node", "request_id": "request", "known_client": "csc",
		"client_user_agent": "csc/1", "user_agent": "bridge/1", "agent_type": "cs-bridge-gateway", "costrict_version": "version",
	}
	for _, empty := range []bool{false, true} {
		var info WorkspaceInfo
		if !empty {
			info = NewWorkspaceInfo(headers)
		} else {
			info = NewWorkspaceInfo(nil)
			for key := range expected {
				expected[key] = ""
			}
		}
		payload, err := json.Marshal(ChatLog{WorkspaceInfo: info})
		if err != nil {
			t.Fatal(err)
		}
		var decoded struct {
			WorkspaceInfo map[string]string `json:"workspace-info"`
		}
		if err := json.Unmarshal(payload, &decoded); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(decoded.WorkspaceInfo, expected) {
			t.Fatalf("workspace-info = %#v, want %#v", decoded.WorkspaceInfo, expected)
		}
	}
}

func TestWorkspaceInfoTaskPriority(t *testing.T) {
	for _, tt := range []struct{ session, legacy, want string }{
		{"session", "legacy", "session"}, {"session", "", "session"},
		{" \t", " legacy ", "legacy"}, {"", "", ""},
	} {
		headers := make(http.Header)
		headers.Set("X-Session-Id", tt.session)
		headers.Set("zgsm-task-id", tt.legacy)
		got := NewWorkspaceInfo(headers)
		if got.TaskID != tt.want || got.SessionID != strings.TrimSpace(tt.session) || got.ZgsmTaskID != strings.TrimSpace(tt.legacy) {
			t.Fatalf("headers %+v: got %+v", tt, got)
		}
	}
}
