package logic

import (
	"testing"
	"time"

	"github.com/zgsm-ai/chat-rag/internal/config"
)

func TestNewChatLogWorkspaceInfo(t *testing.T) {
	logic, _ := setupTestLogic(t, &config.Config{}, nil, "model", nil, &mockResponseWriter{})
	logic.identity.TaskID = "legacy"
	logic.headers.Set("X-Session-Id", "session")
	logic.headers.Set("zgsm-task-id", "legacy")
	logic.headers.Set("X-Project-Id", "project")
	log := logic.newChatLog(time.Now())
	if log.WorkspaceInfo.TaskID != "session" || log.WorkspaceInfo.ProjectID != "project" || log.Identity.TaskID != "legacy" {
		t.Fatalf("unexpected log identity: %+v / %+v", log.Identity, log.WorkspaceInfo)
	}
	logic.headers.Set("X-Project-Id", "changed")
	if log.WorkspaceInfo.ProjectID != "project" {
		t.Fatal("log must snapshot the headers")
	}
	logic.headers = nil
	log = logic.newChatLog(time.Now())
	if log.WorkspaceInfo.TaskID != "" || log.WorkspaceInfo.ProjectPath != "" || log.Identity.TaskID != "legacy" {
		t.Fatal("missing headers must not fall back to legacy identity")
	}
}
