package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zgsm-ai/oidc-auth/internal/repository"
	"github.com/zgsm-ai/oidc-auth/pkg/log"
)

const maxGithubWebhookBodySize = 1 << 20

type githubStarWebhookPayload struct {
	Action     string `json:"action"`
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
	Sender struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
	} `json:"sender"`
}

func (s *Server) githubWebhookHandler(c *gin.Context) {
	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, maxGithubWebhookBodySize))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook body"})
		return
	}

	if !validGithubWebhookSignature(body, c.GetHeader("X-Hub-Signature-256"), s.GitHubWebhookSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid webhook signature"})
		return
	}

	event := c.GetHeader("X-GitHub-Event")
	if event == "ping" {
		c.Status(http.StatusNoContent)
		return
	}
	if event != "star" {
		c.Status(http.StatusNoContent)
		return
	}

	var payload githubStarWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook payload"})
		return
	}

	if !strings.EqualFold(payload.Repository.Owner.Login, s.GitHubOwner) ||
		!strings.EqualFold(payload.Repository.Name, s.GitHubRepo) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unexpected repository"})
		return
	}
	if payload.Sender.ID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing sender id"})
		return
	}

	var starred bool
	switch payload.Action {
	case "created":
		starred = true
	case "deleted":
		starred = false
	default:
		c.Status(http.StatusNoContent)
		return
	}

	repositoryName := fmt.Sprintf("%s.%s", s.GitHubOwner, s.GitHubRepo)
	updateGithubStar := s.updateGithubStar
	if updateGithubStar == nil {
		updateGithubStar = repository.GetDB().UpdateGithubStar
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	rows, err := updateGithubStar(ctx, strconv.FormatInt(payload.Sender.ID, 10), repositoryName, starred)
	if err != nil {
		log.Error(c.Request.Context(), "Failed to process GitHub star webhook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update star status"})
		return
	}

	log.Info(c.Request.Context(), "Processed GitHub star webhook: action=%s github_user=%s rows=%d",
		payload.Action, payload.Sender.Login, rows)
	c.Status(http.StatusNoContent)
}

func validGithubWebhookSignature(body []byte, signature, secret string) bool {
	if secret == "" || !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	provided, err := hex.DecodeString(strings.TrimPrefix(signature, "sha256="))
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hmac.Equal(provided, mac.Sum(nil))
}
