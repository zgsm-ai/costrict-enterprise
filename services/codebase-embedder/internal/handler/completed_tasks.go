package handler

import (
	"fmt"
	"net/http"

	"github.com/zgsm-ai/codebase-indexer/internal/logic"
	"github.com/zgsm-ai/codebase-indexer/internal/response"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
)

// completedTasksHandler 已完成任务处理器
type completedTasksHandler struct {
	svcCtx *svc.ServiceContext
}

// CompletedTasksHandler 创建已完成任务处理器
func CompletedTasksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := &completedTasksHandler{svcCtx: svcCtx}
		handler.ServeHTTP(w, r)
	}
}

// ServeHTTP 处理HTTP请求
func (h *completedTasksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 验证请求方法
	if r.Method != http.MethodGet {
		response.Error(w, fmt.Errorf("method not allowed"))
		return
	}

	// 查询已完成任务
	completedTasksLogic := logic.NewCompletedTasksLogic(r.Context(), h.svcCtx)
	resp, err := completedTasksLogic.GetCompletedTasks()
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Json(w, resp)
}
