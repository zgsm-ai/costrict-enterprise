package helper

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zgsm-ai/chat-rag/internal/model"
)

func TestGetIdentityFromHeadersTaskID(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{name: "missing"},
		{name: "legacy task", headers: map[string]string{"zgsm-task-id": "task-1"}, want: "task-1"},
		{name: "agent session", headers: map[string]string{"X-Session-Id": "session-1"}, want: "session-1"},
		{name: "case insensitive", headers: map[string]string{"x-session-id": "session-1"}, want: "session-1"},
		{name: "legacy takes precedence", headers: map[string]string{"zgsm-task-id": "task-1", "X-Session-Id": "session-1"}, want: "task-1"},
		{name: "empty task falls back", headers: map[string]string{"zgsm-task-id": "", "X-Session-Id": "session-1"}, want: "session-1"},
		{name: "blank task falls back", headers: map[string]string{"zgsm-task-id": " \t", "X-Session-Id": " session-1 "}, want: "session-1"},
		{name: "trim task", headers: map[string]string{"zgsm-task-id": " task-1 "}, want: "task-1"},
		{name: "blank session", headers: map[string]string{"X-Session-Id": " \t"}},
		{name: "request ID is not task ID", headers: map[string]string{"x-request-id": "request-1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)
			for name, value := range tt.headers {
				c.Request.Header.Set(name, value)
			}
			identity := GetIdentityFromHeaders(c)
			if identity.TaskID != tt.want {
				t.Fatalf("TaskID = %q, want %q", identity.TaskID, tt.want)
			}
			payload, err := json.Marshal(model.ChatLog{Identity: *identity})
			if err != nil {
				t.Fatal(err)
			}
			var logged struct {
				Identity struct {
					TaskID string `json:"task_id"`
				} `json:"identity"`
			}
			if err := json.Unmarshal(payload, &logged); err != nil {
				t.Fatal(err)
			}
			if logged.Identity.TaskID != tt.want {
				t.Fatalf("logged task_id = %q, want %q", logged.Identity.TaskID, tt.want)
			}
		})
	}
}
