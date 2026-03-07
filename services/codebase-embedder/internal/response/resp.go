package response

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

const (
	CodeOK = 0

	MessageOk = "ok"

	CodeError = -1
)

type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
}

func Ok(w http.ResponseWriter) {
	httpx.OkJson(w, wrapResponse(nil))
}

func Error(w http.ResponseWriter, e error) {
	logx.WithCallerSkip(2).Errorf("response error: %v", e)
	httpx.WriteJson(w, http.StatusBadRequest, wrapResponse(e)) // TODO 500会触发go-zero breaker
}

func RateLimit(w http.ResponseWriter, e error) {
	logx.WithCallerSkip(2).Errorf("rate limit error: %v", e)
	httpx.WriteJson(w, http.StatusTooManyRequests, wrapResponse(e))
}

func Bytes(w http.ResponseWriter, v []byte) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(v)
	w.Header().Set("Content-Type", "application/octet-stream")
}

func JsonCtx(ctx context.Context, w http.ResponseWriter, v any) {
	httpx.OkJsonCtx(ctx, w, wrapResponse(v))
}

func Json(w http.ResponseWriter, v any) {
	httpx.OkJson(w, wrapResponse(v))
}

func wrapResponse(v any) Response[any] {
	var resp Response[any]
	switch data := v.(type) {
	case *codeMsg:
		resp.Code = data.Code
		resp.Message = data.Message
	case codeMsg:
		resp.Code = data.Code
		resp.Message = data.Message
	case error:
		resp.Code = CodeError
		resp.Message = data.Error()
	default:
		resp.Code = CodeOK
		resp.Message = MessageOk
		resp.Success = true
		resp.Data = v
	}

	return resp
}
