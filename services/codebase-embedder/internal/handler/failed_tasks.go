package handler

import (
	"fmt"
	"net/http"

	"github.com/zgsm-ai/codebase-indexer/internal/logic"
	"github.com/zgsm-ai/codebase-indexer/internal/response"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
)

// failedTasksHandler 失败任务处理器
type failedTasksHandler struct {
	svcCtx *svc.ServiceContext
}

// FailedTasksHandler 创建失败任务处理器
func FailedTasksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := &failedTasksHandler{svcCtx: svcCtx}
		handler.ServeHTTP(w, r)
	}
}

// ServeHTTP 处理HTTP请求
func (h *failedTasksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 验证请求方法
	if r.Method != http.MethodGet {
		response.Error(w, fmt.Errorf("method not allowed"))
		return
	}

	// 查询失败任务
	failedTasksLogic := logic.NewFailedTasksLogic(r.Context(), h.svcCtx)
	resp, err := failedTasksLogic.GetFailedTasks()
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Json(w, resp)
}
