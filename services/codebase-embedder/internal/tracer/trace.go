package tracer

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
	"strings"
)

const Key = "trace_id"

func WithTrace(ctx context.Context) logx.Logger {
	logger := logx.WithContext(ctx)
	taskTraceId, ok := ctx.Value(Key).(string)
	if ok && taskTraceId != types.EmptyString {
		logger = logger.WithFields(logx.Field(Key, taskTraceId))
	}
	return logger
}

func RequestTraceId(baseId int) string {
	return fmt.Sprintf("req-%d@%s", baseId, strings.ReplaceAll(uuid.New().String(), "-", types.EmptyString))
}

func MsgTraceId(baseId string) string {
	return fmt.Sprintf("msg-%s", baseId)
}

func TaskTraceId(baseId int) string {
	return fmt.Sprintf("task-%d@%s", baseId, strings.ReplaceAll(uuid.New().String(), "-", types.EmptyString))
}
