package model

import (
	"code-completion/pkg/config"
	"code-completion/pkg/tokenizers"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type OpenAIModelManager struct {
	models []LLM
	mutex  sync.Mutex
	index  int
}

type NewLLM func(*config.ModelConfig, *tokenizers.Tokenizer) LLM

var modelDefs = map[string]NewLLM{
	"openai":   NewOpenAIModel,
	"deepseek": NewOpenAIModel,
}

func GetAutoModel() LLM {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	modelLen := len(manager.models)
	if modelLen == 0 {
		panic(manager)
	}
	// 采用轮转法选择模型进行响应
	var model LLM
	if manager.index < modelLen {
		model = manager.models[manager.index]
		manager.index++
	} else {
		manager.index = 1
		model = manager.models[0]
	}
	return model
}

func GetModel(idx int) LLM {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if idx >= len(manager.models) {
		panic(manager)
	}
	return manager.models[idx]
}

var manager = &OpenAIModelManager{}

func Init(cfgModels []config.ModelConfig) error {
	models := make([]LLM, 0)
	for _, c := range cfgModels {
		token, err := tokenizers.NewTokenizer(c.TokenizerPath)
		if err != nil {
			zap.L().Error("init tokenizer error", zap.String("tokenizerPath", c.TokenizerPath), zap.Error(err))
			continue
		}
		newLLM, exists := modelDefs[c.Provider]
		if !exists {
			newLLM = NewOpenAIModel
		}
		models = append(models, newLLM(&c, token))
	}
	if len(models) == 0 {
		zap.L().Fatal("No models available")
		return fmt.Errorf("no models available")
	}
	manager.models = models
	return nil
}
