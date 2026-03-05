package model

import (
	"bytes"
	"completion-agent/pkg/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OpenAICompletion struct {
	cfg    *config.ModelConfig
	client *http.Client
}

func NewOpenAICompletion(c *config.ModelConfig) LLM {
	return &OpenAICompletion{
		cfg: c,
		client: &http.Client{
			Timeout: c.Timeout.Duration(),
		},
	}
}

func (m *OpenAICompletion) Config() *config.ModelConfig {
	return m.cfg
}

/**
 * 获取加了FIM标记的prompt文本
 */
func (m *OpenAICompletion) getFimPrompt(prefix, suffix, codeContext string, cfg *config.ModelConfig) string {
	return cfg.FimBegin + codeContext + "\n" + prefix + cfg.FimHole + suffix + cfg.FimEnd
}

func (m *OpenAICompletion) Completions(ctx context.Context, p *CompletionParameter) (*CompletionResponse, CompletionStatus, error) {
	var prefix string
	if m.cfg.FimMode {
		prefix = m.getFimPrompt(p.Prefix, p.Suffix, p.CodeContext, m.cfg)
	} else {
		if p.CodeContext != "" {
			prefix = strings.Join([]string{p.CodeContext, p.Prefix}, "\n")
		} else {
			prefix = p.Prefix
		}
	}
	maxTokens := min(p.MaxTokens, m.cfg.MaxOutput)
	data := map[string]interface{}{
		"model":       m.cfg.ModelName,
		"prompt":      prefix,
		"stop":        p.Stop,
		"temperature": p.Temperature,
		"max_tokens":  maxTokens,
		"stream":      false,
	}
	if !m.cfg.FimMode && p.Suffix != "" {
		data["suffix"] = p.Suffix
	}
	// 将data转换为JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, StatusServerError, err
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", m.cfg.CompletionsUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, StatusReqError, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", m.cfg.Authorization)

	// 发送请求
	resp, err := m.client.Do(req)
	if err != nil {
		status := StatusServerError
		switch err {
		case context.Canceled:
			status = StatusCanceled
		case context.DeadlineExceeded:
			status = StatusTimeout
		}
		return nil, status, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, StatusServerError, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, StatusModelError, fmt.Errorf("invalid StatusCode(%d)", resp.StatusCode)
	}
	var rsp CompletionResponse
	if err := json.Unmarshal(body, &rsp); err != nil {
		return nil, StatusServerError, err
	}
	return &rsp, StatusSuccess, nil
}
