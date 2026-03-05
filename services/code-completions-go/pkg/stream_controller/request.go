package stream_controller

import (
	"code-completion/pkg/completions"
	"code-completion/pkg/model"
	"context"
	"strings"
)

// 客户端请求包装器
type ClientRequest struct {
	Para     *model.CompletionParameter           // 补全请求参数
	Perf     *completions.CompletionPerformance   // 性能统计
	Canceled bool                                 // 请求是否被取消
	ctx      context.Context                      // 请求关联的协程上下文
	cancel   context.CancelFunc                   // 可以取消执行请求的协程
	rspChan  chan *completions.CompletionResponse // 响应通道
}

func (r *ClientRequest) GetDetails() map[string]interface{} {
	var linePrefix, lineSuffix string
	lines := strings.Split(r.Para.Prefix, "\n")
	if len(lines) > 0 {
		linePrefix = lines[len(lines)-1]
	}
	lines = strings.Split(r.Para.Suffix, "\n")
	if len(lines) > 0 {
		lineSuffix = lines[0]
		if len(lines) > 1 {
			lineSuffix += "\n"
		}
	}
	return map[string]interface{}{
		"completion_id": r.Para.CompletionID,
		"client_id":     r.Para.ClientID,
		"model":         r.Para.Model,
		"prompt": map[string]interface{}{
			"prefix":      len(r.Para.Prefix),
			"suffix":      len(r.Para.Suffix),
			"context":     len(r.Para.CodeContext),
			"total":       len(r.Para.Prefix) + len(r.Para.Suffix) + len(r.Para.CodeContext),
			"line_prefix": linePrefix,
			"line_suffix": lineSuffix,
		},
		"performance": r.Perf,
		"canceled":    r.Canceled,
	}
}

func (r *ClientRequest) GetSummary() map[string]interface{} {
	return map[string]interface{}{
		"completion_id": r.Para.CompletionID,
		"client_id":     r.Para.ClientID,
		"prompt":        len(r.Para.Prefix) + len(r.Para.Suffix) + len(r.Para.CodeContext),
		"canceled":      r.Canceled,
	}
}
