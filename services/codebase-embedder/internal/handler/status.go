package handler

import (
	"net/http"

	"github.com/zgsm-ai/codebase-indexer/internal/logic"
	"github.com/zgsm-ai/codebase-indexer/internal/response"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// statusHandler 文件状态处理器
type statusHandler struct {
	svcCtx *svc.ServiceContext
}

// StatusHandler 创建文件状态处理器
func StatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := &statusHandler{svcCtx: svcCtx}
		handler.ServeHTTP(w, r)
	}
}

// ServeHTTP 处理HTTP请求
func (h *statusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req types.FileStatusRequest
	if err := httpx.Parse(r, &req); err != nil {
		response.Error(w, err)
		return
	}

	// 查询文件处理状态
	statusLogic := logic.NewStatusLogic(r.Context(), h.svcCtx)
	statusResp, err := statusLogic.GetFileStatus(&req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.Json(w, statusResp)
}