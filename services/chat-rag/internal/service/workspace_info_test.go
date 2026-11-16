package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/zgsm-ai/chat-rag/internal/model"
)

func TestWorkspaceInfoReportPayload(t *testing.T) {
	var received map[string]json.RawMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Error(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	reporter := NewChatMetricsReporter(server.URL, http.MethodPost, "")
	headers := make(http.Header)
	headers.Set("X-Session-Id", "session")
	headers.Set("zgsm-task-id", "legacy")
	headers.Set("X-Project-Id", "project")
	for _, info := range []model.WorkspaceInfo{model.NewWorkspaceInfo(headers), {}} {
		log := &model.ChatLog{WorkspaceInfo: info, Identity: model.Identity{TaskID: "legacy"}}
		report := reporter.convertChatLogToReport(log, nil)
		if report.Label.TaskID != "legacy" {
			t.Fatal("legacy label changed")
		}
		if err := reporter.sendReport(report, ""); err != nil {
			t.Fatal(err)
		}
		logJSON, err := json.Marshal(log)
		if err != nil {
			t.Fatal(err)
		}
		var logged map[string]json.RawMessage
		if err := json.Unmarshal(logJSON, &logged); err != nil {
			t.Fatal(err)
		}
		var got, want map[string]string
		if err := json.Unmarshal(received["workspace-info"], &got); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(logged["workspace-info"], &want); err != nil {
			t.Fatal(err)
		}
		if len(got) != 16 || !reflect.DeepEqual(got, want) {
			t.Fatalf("report workspace-info = %#v, log = %#v", got, want)
		}
	}
}
