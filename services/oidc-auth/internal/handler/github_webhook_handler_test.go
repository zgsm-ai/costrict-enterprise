package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestValidGithubWebhookSignature(t *testing.T) {
	body := []byte(`{"action":"created"}`)
	secret := "webhook-secret"
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name      string
		body      []byte
		signature string
		secret    string
		want      bool
	}{
		{name: "valid", body: body, signature: signature, secret: secret, want: true},
		{name: "modified body", body: []byte(`{"action":"deleted"}`), signature: signature, secret: secret},
		{name: "invalid encoding", body: body, signature: "sha256=not-hex", secret: secret},
		{name: "wrong algorithm", body: body, signature: "sha1=abc", secret: secret},
		{name: "empty secret", body: body, signature: signature},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validGithubWebhookSignature(tt.body, tt.signature, tt.secret); got != tt.want {
				t.Fatalf("validGithubWebhookSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGithubWebhookHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "webhook-secret"

	tests := []struct {
		name           string
		event          string
		body           string
		signature      string
		wantStatus     int
		wantGithubID   string
		wantRepository string
		wantStarred    bool
		wantCalls      int
	}{
		{
			name: "star created", event: "star",
			body:       `{"action":"created","repository":{"name":"costrict","owner":{"login":"zgsm-ai"}},"sender":{"id":123,"login":"alice"}}`,
			wantStatus: http.StatusNoContent, wantGithubID: "123", wantRepository: "zgsm-ai.costrict", wantStarred: true, wantCalls: 1,
		},
		{
			name: "star deleted", event: "star",
			body:       `{"action":"deleted","repository":{"name":"costrict","owner":{"login":"zgsm-ai"}},"sender":{"id":456,"login":"bob"}}`,
			wantStatus: http.StatusNoContent, wantGithubID: "456", wantRepository: "zgsm-ai.costrict", wantCalls: 1,
		},
		{
			name: "repository mismatch", event: "star",
			body:       `{"action":"created","repository":{"name":"other","owner":{"login":"zgsm-ai"}},"sender":{"id":123}}`,
			wantStatus: http.StatusBadRequest,
		},
		{name: "unsupported event", event: "issues", body: `{}`, wantStatus: http.StatusNoContent},
		{name: "invalid signature", event: "star", body: `{}`, signature: "sha256=00", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			server := Server{
				GitHubWebhookSecret: secret,
				GitHubOwner:         "zgsm-ai",
				GitHubRepo:          "costrict",
				updateGithubStar: func(_ context.Context, githubID, repository string, starred bool) (int64, error) {
					calls++
					if githubID != tt.wantGithubID || repository != tt.wantRepository || starred != tt.wantStarred {
						return 0, fmt.Errorf("unexpected update: id=%s repository=%s starred=%v", githubID, repository, starred)
					}
					return 1, nil
				},
			}

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/oidc-auth/api/v1/webhooks/github", bytes.NewBufferString(tt.body))
			request.Header.Set("X-GitHub-Event", tt.event)
			signature := tt.signature
			if signature == "" {
				signature = signGithubWebhook([]byte(tt.body), secret)
			}
			request.Header.Set("X-Hub-Signature-256", signature)

			router := gin.New()
			router.POST("/oidc-auth/api/v1/webhooks/github", server.githubWebhookHandler)
			router.ServeHTTP(recorder, request)
			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			if calls != tt.wantCalls {
				t.Fatalf("update calls = %d, want %d", calls, tt.wantCalls)
			}
		})
	}
}

func signGithubWebhook(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
