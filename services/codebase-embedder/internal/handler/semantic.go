package handler

import (
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zgsm-ai/codebase-indexer/internal/logic"
	"github.com/zgsm-ai/codebase-indexer/internal/response"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

func semanticSearchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SemanticSearchRequest
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

		// 打印请求参数用于调试
		fmt.Printf("Semantic search request received - ClientId: %s, CodebasePath: %s, Query: %s, TopK: %d, ScoreThreshold: %f\n",
			req.ClientId, req.CodebasePath, req.Query, req.TopK, req.ScoreThreshold)

		l := logic.NewSemanticSearchLogic(r.Context(), svcCtx)
		resp, err := l.SemanticSearch(&req, authorization)
		if err != nil {
			response.Error(w, err)
		} else {
			response.Json(w, resp)
		}
	}
}
