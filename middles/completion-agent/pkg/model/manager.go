package model

import (
	"completion-agent/pkg/config"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

/**
 * OpenAI模型管理器结构体
 * @description
 * - 管理多个LLM模型实例
 * - 提供模型轮询选择机制
 * - 使用互斥锁确保线程安全
 * - 维护当前模型索引用于轮询
 * @example
 * // 通常通过Init函数初始化
 * model := GetAutoModel()
 * response, err := model.Completions(ctx, &para)
 */
type LLManager struct {
	models []LLM
	mutex  sync.Mutex
	index  int
}

/**
 * 模型工厂函数类型
 * @param {*config.ModelConfig} - 模型配置参数，包含模型名称、URL等信息
 * @returns {LLM} 返回初始化好的LLM接口实例
 * @description
 * - 定义创建新LLM实例的函数类型
 * - 接收模型配置作为参数
 * - 返回实现LLM接口的实例
 * - 用于支持多种模型类型的创建
 * @example
 * var factory NewLLM = NewOpenAICompletion
 * model := factory(&config.ModelConfig{ModelName: "gpt-3.5-turbo"})
 */
type NewLLM func(*config.ModelConfig) LLM

var modelDefs = map[string]NewLLM{
	"openai":  NewOpenAICompletion,
	"sangfor": NewSangforCompletion,
}

/**
 * 自动获取模型实例
 * @returns {LLM} 返回选中的LLM模型实例
 * @description
 * - 使用轮询算法自动选择模型
 * - 线程安全，使用互斥锁保护共享状态
 * - 如果没有可用模型会panic
 * - 按顺序循环使用所有配置的模型
 * @throws
 * - 如果没有可用模型，会导致程序panic
 * @example
 * model := GetAutoModel()
 * response, err := model.Completions(ctx, &para)
 */
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

/**
 * 根据索引获取模型实例
 * @param {int} idx - 模型索引，从0开始
 * @returns {LLM} 返回指定索引的LLM模型实例
 * @description
 * - 根据指定的索引获取模型实例
 * - 线程安全，使用互斥锁保护共享状态
 * - 如果索引超出范围会panic
 * - 用于获取特定的模型实例
 * @throws
 * - 如果索引超出范围，会导致程序panic
 * @example
 * model := GetModel(0)
 * response, err := model.Completions(ctx, &para)
 */
func GetModel(idx int) LLM {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if idx >= len(manager.models) {
		panic(manager)
	}
	return manager.models[idx]
}

var manager = &LLManager{}

/**
 * 初始化模型管理器
 * @param {[]config.ModelConfig} cfgModels - 模型配置数组，包含所有要初始化的模型配置
 * @returns {error} 返回初始化过程中的错误，成功返回nil
 * @description
 * - 根据配置数组初始化所有模型实例
 * - 根据provider类型选择对应的模型工厂函数
 * - 如果provider不存在，默认使用Sangfor模型
 * - 如果没有可用模型，记录fatal日志并返回错误
 * - 线程安全，初始化完成后可用于模型选择
 * @throws
 * - 如果没有可用模型，记录fatal日志并返回错误
 * @example
 * models := []config.ModelConfig{
 *     {Provider: "openai", ModelName: "gpt-3.5-turbo"},
 *     {Provider: "sangfor", ModelName: "model-v1"}
 * }
 * if err := Init(models); err != nil {
 *     log.Fatal("模型初始化失败:", err)
 * }
 */
func Init(cfgModels []config.ModelConfig) error {
	models := make([]LLM, 0)
	for _, c := range cfgModels {
		newLLM, exists := modelDefs[c.Provider]
		if !exists {
			newLLM = NewSangforCompletion
		}
		models = append(models, newLLM(&c))
	}
	if len(models) == 0 {
		zap.L().Fatal("No models available")
		return fmt.Errorf("no models available")
	}
	manager.models = models
	return nil
}
