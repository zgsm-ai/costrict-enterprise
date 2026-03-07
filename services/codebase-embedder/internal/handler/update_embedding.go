package handler

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zgsm-ai/codebase-indexer/internal/logic"
	"github.com/zgsm-ai/codebase-indexer/internal/response"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

func updateEmbeddingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateEmbeddingPathRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Error(w, err)
			return
		}

		// 验证必填字段
		if req.ClientId == "" {
			response.Error(w, errors.New("missing required parameter: clientId"))
			return
		}
		if req.CodebasePath == "" {
			response.Error(w, errors.New("missing required parameter: codebasePath"))
			return
		}
		if req.OldPath == "" {
			response.Error(w, errors.New("missing required parameter: oldPath"))
			return
		}
		if req.NewPath == "" {
			response.Error(w, errors.New("missing required parameter: newPath"))
			return
		}

		// 规范化路径，确保使用正斜杠
		req.OldPath = filepath.ToSlash(req.OldPath)
		req.NewPath = filepath.ToSlash(req.NewPath)

		l := logic.NewUpdateEmbeddingLogic(r.Context(), svcCtx)
		resp, err := l.UpdateEmbeddingPath(&req)
		if err != nil {
			response.Error(w, err)
		} else {
			response.Json(w, resp)
		}
	}
}
