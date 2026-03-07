package handler

import (
	"context"
	"net/http"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/zgsm-ai/codebase-indexer/internal/logic"
	"github.com/zgsm-ai/codebase-indexer/internal/response"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

// QueryCodebaseHandler 查询代码库处理器
type QueryCodebaseHandler struct {
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewQueryCodebaseHandler 创建查询代码库处理器
func NewQueryCodebaseHandler(svcCtx *svc.ServiceContext) *QueryCodebaseHandler {
	return &QueryCodebaseHandler{
		svcCtx: svcCtx,
		Logger: logx.WithContext(context.Background()),
	}
}

// QueryCodebase 查询代码库接口
func (h *QueryCodebaseHandler) QueryCodebase(w http.ResponseWriter, r *http.Request) {
	// 1. 解析和验证请求参数
	var req types.CodebaseQueryRequest
	if err := httpx.Parse(r, &req); err != nil {
		h.Logger.Errorf("解析请求参数失败, error: %v, request: %+v", err, r)
		h.Logger.Errorf("请求详情 - Method: %s, URL: %s, Content-Type: %s", r.Method, r.URL.String(), r.Header.Get("Content-Type"))
		body := make([]byte, 1024)
		if n, err := r.Body.Read(body); err == nil || err.Error() == "EOF" {
			h.Logger.Errorf("请求体内容: %s", string(body[:n]))
		}
		response.Json(w, response.NewParamError("请求参数解析失败"))
		return
	}

	// 2. 验证请求参数
	if err := h.validateRequest(&req); err != nil {
		h.Logger.Errorf("验证请求参数失败, req: %+v, error: %v", req, err)
		response.Json(w, err)
		return
	}

	// 5. 调用业务逻辑层
	queryLogic := logic.NewQueryCodebaseLogic(r.Context(), h.svcCtx)
	resp, err := queryLogic.QueryCodebase(&req)
	if err != nil {
		h.Logger.Errorf("查询代码库失败, req: %+v, error: %v", req, err)
		response.Json(w, err)
		return
	}

	// 6. 返回成功响应
	response.Json(w, resp)
}

// validateRequest 验证请求参数
func (h *QueryCodebaseHandler) validateRequest(req *types.CodebaseQueryRequest) error {
	if req.ClientId == "" {
		return response.NewParamError("clientId不能为空")
	}
	if req.CodebasePath == "" {
		return response.NewParamError("codebasePath不能为空")
	}
	if req.CodebaseName == "" {
		return response.NewParamError("codebaseName不能为空")
	}
	return nil
}