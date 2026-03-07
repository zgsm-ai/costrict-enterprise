package handler

import (
	"log"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zgsm-ai/codebase-indexer/internal/logic"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

type CodebaseTreeHandler struct {
	svcCtx *svc.ServiceContext
}

func NewCodebaseTreeHandler(svcCtx *svc.ServiceContext) *CodebaseTreeHandler {
	return &CodebaseTreeHandler{
		svcCtx: svcCtx,
	}
}

func (h *CodebaseTreeHandler) TreeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[DEBUG] CodebaseTreeHandler.TreeHandler 被调用，方法:", r.Method, "路径:", r.URL.Path)
	var req types.CodebaseTreeRequest
	if err := httpx.Parse(r, &req); err != nil {
		log.Println("[DEBUG] 解析请求失败:", err)
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	log.Println("[DEBUG] 解析请求成功，clientId:", req.ClientId, "codebasePath:", req.CodebasePath)
	l := logic.NewCodebaseTreeLogic(r.Context(), h.svcCtx)
	resp, err := l.GetCodebaseTree(&req)
	if err != nil {
		log.Println("[DEBUG] 处理请求失败:", err)
		httpx.ErrorCtx(r.Context(), w, err)
	} else {
		log.Println("[DEBUG] 处理请求成功")
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
