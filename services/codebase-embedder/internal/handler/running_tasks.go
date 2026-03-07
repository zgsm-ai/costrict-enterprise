package handler

import (
	"fmt"
	"net/http"

	"github.com/zgsm-ai/codebase-indexer/internal/logic"
	"github.com/zgsm-ai/codebase-indexer/internal/response"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
)

// runningTasksHandler 运行中任务处理器
type runningTasksHandler struct {
	svcCtx *svc.ServiceContext
}

// RunningTasksHandler 创建运行中任务处理器
func RunningTasksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := &runningTasksHandler{svcCtx: svcCtx}
		handler.ServeHTTP(w, r)
	}
}

// ServeHTTP 处理HTTP请求
func (h *runningTasksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 验证请求方法
	if r.Method != http.MethodGet {
		response.Error(w, fmt.Errorf("method not allowed"))
		return
	}

	// 查询运行中任务
	runningTasksLogic := logic.NewRunningTasksLogic(r.Context(), h.svcCtx)
	resp, err := runningTasksLogic.GetRunningTasks()
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Json(w, resp)
}
