package log

import (
	"fmt"

	"go.uber.org/zap"
)

type GormWriter struct {
	logger *zap.Logger
}

func NewGormWriter() *GormWriter {
	return &GormWriter{
		logger: GetLogger().Named("gorm"),
	}
}

func (w *GormWriter) Printf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	w.logger.Info(msg)
}
