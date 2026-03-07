package handler

import (
	"net/http"

	"github.com/zgsm-ai/codebase-indexer/internal/response"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zgsm-ai/codebase-indexer/internal/logic"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

func documentsSummaryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DocumentsSummaryRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Error(w, err)
			return
		}

		l := logic.NewDocumentsSummaryLogic(r.Context(), svcCtx)
		resp, err := l.DocumentsSummary(&req)
		if err != nil {
			response.Error(w, err)
		} else {
			response.Json(w, resp)
		}
	}
}
