package handler

import (
	"net/http"

	"github.com/zgsm-ai/codebase-indexer/internal/response"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zgsm-ai/codebase-indexer/internal/logic"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

func combinedSummaryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CombinedSummaryRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Error(w, err)
			return
		}

		// 从请求头获取 Authorization
		authorization := r.Header.Get("Authorization")

		// 验证 Authorization 头是否存在
		if authorization == "" {
			response.Error(w, response.NewAuthError("missing Authorization header"))
			return
		}

		l := logic.NewCombinedSummaryLogic(r.Context(), svcCtx)
		resp, err := l.CombinedSummary(&req, authorization)
		if err != nil {
			response.Error(w, err)
		} else {
			response.Json(w, resp)
		}
	}
}
