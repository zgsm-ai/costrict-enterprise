package tokenizers

import (
	"completion-agent/pkg/config"

	"go.uber.org/zap"
)

var global *Tokenizer

func Init() error {
	t, err := NewTokenizer(config.Wrapper.Tokenizer.Path)
	if err != nil {
		zap.L().Error("init tokenizer error",
			zap.String("path", config.Wrapper.Tokenizer.Path), zap.Error(err))
		return err
	}
	global = t
	return nil
}

func GetTokenizer() *Tokenizer {
	return global
}
