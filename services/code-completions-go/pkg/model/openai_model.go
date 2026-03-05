package model

import (
	"bytes"
	"code-completion/pkg/config"
	"code-completion/pkg/tokenizers"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OpenAIModel struct {
	cfg       *config.ModelConfig
	tokenizer *tokenizers.Tokenizer
}

func NewOpenAIModel(c *config.ModelConfig, t *tokenizers.Tokenizer) LLM {
	return &OpenAIModel{
		cfg:       c,
		tokenizer: t,
	}
}

func (m *OpenAIModel) Config() *config.ModelConfig {
	return m.cfg
}

func (m *OpenAIModel) Tokenizer() *tokenizers.Tokenizer {
	return m.tokenizer
}

/**
 * 获取加了FIM标记的prompt文本
 * @param {string} prefix - 代码前缀文本
 * @param {string} suffix - 代码后缀文本
 * @param {string} codeContext - 代码上下文文本
 * @param {*config.ModelConfig} cfg - 模型配置，包含FIM相关标记
 * @returns {string} 返回添加了FIM标记的完整prompt文本
 * @description
 * - 按照FIM(Fill In the Middle)格式组装prompt
 * - 使用配置中的FIM标记：FimBegin、FimHole、FimEnd
 * - 格式为：FimBegin + codeContext + "\n" + prefix + FimHole + suffix + FimEnd
 * - 用于支持FIM模式的代码补全
 * @example
 * cfg := &config.ModelConfig{
 *     FimBegin: "<fim-prefix>",
 *     FimHole: "<fim-suffix>",
 *     FimEnd: "<fim-middle>",
 * }
 * prompt := handler.getFimPrompt("function test", "}", "context", cfg)
 * // prompt = "<fim-prefix>context\nfunction test<fim-suffix>}<fim-middle>"
 */
func (m *OpenAIModel) getFimPrompt(prefix, suffix, codeContext string, cfg *config.ModelConfig) string {
	return cfg.FimBegin + codeContext + "\n" + prefix + cfg.FimHole + suffix + cfg.FimEnd
}

func (m *OpenAIModel) Completions(ctx context.Context, p *CompletionParameter) (*CompletionResponse, *CompletionVerbose, CompletionStatus, error) {
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
	var verbose CompletionVerbose
	verbose.Id = m.cfg.ModelTitle
	verbose.Input = data
	// 将data转换为JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, &verbose, StatusServerError, err
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", m.cfg.CompletionsUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, &verbose, StatusReqError, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", m.cfg.Authorization)

	// 发送请求
	client := &http.Client{
		Timeout: m.cfg.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		status := StatusServerError
		switch err {
		case context.Canceled:
			status = StatusCanceled
		case context.DeadlineExceeded:
			status = StatusTimeout
		}
		return nil, &verbose, status, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &verbose, StatusServerError, err
	}
	json.Unmarshal(body, &verbose.Output)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &verbose, StatusModelError, fmt.Errorf("Invalid StatusCode(%d)", resp.StatusCode)
	}
	var rsp CompletionResponse
	if err := json.Unmarshal(body, &rsp); err != nil {
		return nil, &verbose, StatusServerError, err
	}
	return &rsp, &verbose, StatusSuccess, nil
}
