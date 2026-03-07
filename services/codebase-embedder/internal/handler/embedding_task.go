package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/zgsm-ai/codebase-indexer/internal/logic"
	"github.com/zgsm-ai/codebase-indexer/internal/response"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

func taskHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.IndexTaskRequest
		// 修改解析逻辑，从form-data解析参数
		if err := r.ParseMultipartForm(32 << 20); err != nil { // 最大32MB
			response.Error(w, err)
			return
		}

		// 手动映射form参数到请求结构体
		req.ClientId = r.FormValue("clientId")
		req.CodebasePath = r.FormValue("codebasePath")
		req.CodebaseName = r.FormValue("codebaseName")
		req.UploadToken = r.FormValue("uploadToken")
		req.ExtraMetadata = r.FormValue("extraMetadata")
		req.FileTotals = 1 // 默认值

		// 解析fileTotals字段
		if fileTotals := r.FormValue("fileTotals"); fileTotals != "" {
			fmt.Sscanf(fileTotals, "%d", &req.FileTotals)
		}

		// 解析并获取RequestId（必填参数）
		req.RequestId = r.Header.Get("X-Request-ID")
		if req.RequestId == "" {
			response.Error(w, errors.New("missing required header: X-Request-ID"))
			return
		}
		fmt.Printf("Received RequestId: %s\n", req.RequestId)

		// 验证必填字段
		if req.ClientId == "" {
			response.Error(w, errors.New("missing required parameter: clientId"))
			return
		}
		if req.CodebasePath == "" {
			response.Error(w, errors.New("missing required parameter: codebasePath"))
			return
		}
		if req.CodebaseName == "" {
			response.Error(w, errors.New("missing required parameter: codebaseName"))
			return
		}

		l := logic.NewTaskLogic(r.Context(), svcCtx)
		resp, err := l.SubmitTask(&req, r)
		if err != nil {
			response.Error(w, err)
		} else {
			response.Json(w, resp)
		}
	}
}
