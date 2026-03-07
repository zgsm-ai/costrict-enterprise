package handler

import (
	"errors"
	"net/http"

	"github.com/zgsm-ai/codebase-indexer/internal/logic"
	"github.com/zgsm-ai/codebase-indexer/internal/response"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// tokenHandler token处理器
type tokenHandler struct {
	svcCtx *svc.ServiceContext
}

// TokenHandler 创建token处理器
func TokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := &tokenHandler{svcCtx: svcCtx}
		handler.ServeHTTP(w, r)
	}
}

// ServeHTTP 处理HTTP请求
func (h *tokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req types.TokenRequest
	if err := httpx.Parse(r, &req); err != nil {
		response.Error(w, err)
		return
	}

	// 创建token逻辑
	tokenLogic := logic.NewTokenLogic(r.Context(), h.svcCtx)
	tokenResp, err := tokenLogic.GenerateToken(&req)
	if err != nil {
		// 检查是否为限流错误
		if errors.Is(err, types.ErrRateLimitReached) {
			response.RateLimit(w, err)
			return
		}
		response.Error(w, err)
		return
	}
	response.Json(w, tokenResp)
}
