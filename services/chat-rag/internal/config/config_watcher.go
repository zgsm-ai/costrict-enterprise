package config

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"go.uber.org/zap"
)

// ConfigChangeHandler 配置变更处理器接口
type ConfigChangeHandler interface {
	// GetDataId 返回配置的数据ID
	GetDataId() string
	// OnChange 处理配置变更
	OnChange(data string) error
	// GetConfig 获取当前缓存的配置
	GetConfig() interface{}
}

// GenericConfigHandler 通用配置处理器
type GenericConfigHandler struct {
	dataId    string
	configPtr interface{}
	mutex     sync.RWMutex
	onChange  func(interface{})
	unmarshal func(string, interface{}) error
}

// NewGenericConfigHandler 创建通用配置处理器
func NewGenericConfigHandler(dataId string, configType interface{}, onChange func(interface{})) *GenericConfigHandler {
	return &GenericConfigHandler{
		dataId:    dataId,
		configPtr: configType,
		onChange:  onChange,
		unmarshal: unmarshalYAMLContent,
	}
}

// GetDataId 返回配置的数据ID
func (h *GenericConfigHandler) GetDataId() string {
	return h.dataId
}

// OnChange 处理配置变更
func (h *GenericConfigHandler) OnChange(data string) error {
	// 创建新的配置实例
	newConfig, err := h.createConfigInstance()
	if err != nil {
		return fmt.Errorf("failed to create config instance: %w", err)
	}

	// 解析YAML内容
	if err := h.unmarshal(data, newConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 更新缓存
	h.mutex.Lock()
	h.configPtr = newConfig
	h.mutex.Unlock()

	// 调用变更回调
	if h.onChange != nil {
		h.onChange(newConfig)
	}

	logger.Info("Configuration updated successfully",
		zap.String("dataId", h.dataId))

	return nil
}

// createConfigInstance 创建配置实例
func (h *GenericConfigHandler) createConfigInstance() (interface{}, error) {
	// 根据现有配置类型创建新实例
	if h.configPtr == nil {
		return nil, fmt.Errorf("config type not initialized")
	}

	// 使用反射创建相同类型的新实例
	configType := reflect.TypeOf(h.configPtr)
	if configType.Kind() == reflect.Ptr {
		configType = configType.Elem()
	}

	return reflect.New(configType).Interface(), nil
}

// GetConfig 获取当前缓存的配置
func (h *GenericConfigHandler) GetConfig() interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.configPtr
}

// parseFlexibleTime parses time string with flexible formats and auto-completion
// Supports:
// - YYYY-MM-DD (auto-completed to YYYY-MM-DDT00:00:00 using server local timezone)
// - YYYY-MM-DDTHH (auto-completed to YYYY-MM-DDTHH:00:00 using server local timezone)
// - YYYY-MM-DDTHH:MM (auto-completed to YYYY-MM-DDTHH:MM:00 using server local timezone)
// - YYYY-MM-DDTHH:MM:SS (parsed using server local timezone)
// - YYYY-MM-DD HH:MM (space separator, same as above)
// - YYYY-MM-DDTHH:MM:SS+ZZ:ZZ or YYYY-MM-DDTHH:MM+ZZ:ZZ (with timezone, used as-is for backward compatibility)
func parseFlexibleTime(timeStr string) (time.Time, error) {
	// Check for empty input
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("time string is empty")
	}

	// Try parsing RFC3339 format (complete format with timezone)
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t, nil
	}

	// Try parsing without seconds but with timezone: YYYY-MM-DDTHH:MM+ZZ:ZZ
	if t, err := time.Parse("2006-01-02T15:04-07:00", timeStr); err == nil {
		return t, nil
	}

	// Auto-completion logic: use server local timezone (time.Local)
	var completeStr string
	separator := " " // Default to space separator for auto-completed formats

	if len(timeStr) == 10 {
		// YYYY-MM-DD
		completeStr = timeStr + " 00:00:00"
	} else if len(timeStr) == 13 {
		// YYYY-MM-DD T HH (separator is either 'T' or ' ')
		if timeStr[10] == 'T' || timeStr[10] == ' ' {
			separator = string(timeStr[10])
			completeStr = timeStr + ":00:00"
		} else {
			return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
		}
	} else if len(timeStr) == 16 {
		// YYYY-MM-DD T HH:MM (separator is either 'T' or ' ')
		if timeStr[10] == 'T' || timeStr[10] == ' ' {
			separator = string(timeStr[10])
			completeStr = timeStr + ":00"
		} else {
			return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
		}
	} else if len(timeStr) == 19 {
		// YYYY-MM-DD T HH:MM:SS (separator is either 'T' or ' ')
		if timeStr[10] == 'T' || timeStr[10] == ' ' {
			separator = string(timeStr[10])
			completeStr = timeStr
		} else {
			return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
		}
	} else {
		return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
	}

	// Set layout based on separator
	layout := "2006-01-02T15:04:05"
	if separator == " " {
		layout = "2006-01-02 15:04:05"
	}

	// Parse the completed string using server local timezone
	return time.ParseInLocation(layout, completeStr, time.Local)
}

// unmarshalYAMLContent 解析YAML内容
func unmarshalYAMLContent(content string, target interface{}) error {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(content)); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Set up decode hook for flexible time parsing
	decodeHook := mapstructure.ComposeDecodeHookFunc(
		// String to time.Time with flexible parsing
		func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
			if f.Kind() != reflect.String {
				return data, nil
			}
			if t != reflect.TypeOf(time.Time{}) {
				return data, nil
			}

			timeStr, ok := data.(string)
			if !ok {
				return data, nil
			}

			return parseFlexibleTime(timeStr)
		},
	)

	if err := v.Unmarshal(target, viper.DecodeHook(decodeHook)); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// ConfigWatcher 配置监听器
type ConfigWatcher struct {
	client      config_client.IConfigClient
	config      NacosConfig
	handlers    map[string]ConfigChangeHandler
	mutex       sync.RWMutex
	isConnected bool
}

// NewConfigWatcher 创建配置监听器
func NewConfigWatcher(config NacosConfig, client config_client.IConfigClient) *ConfigWatcher {
	return &ConfigWatcher{
		client:      client,
		config:      config,
		handlers:    make(map[string]ConfigChangeHandler),
		isConnected: client != nil,
	}
}

// RegisterHandler 注册配置变更处理器
func (w *ConfigWatcher) RegisterHandler(handler ConfigChangeHandler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	dataId := handler.GetDataId()
	if dataId == "" {
		return fmt.Errorf("dataId cannot be empty")
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if _, exists := w.handlers[dataId]; exists {
		return fmt.Errorf("handler for dataId %s already registered", dataId)
	}

	w.handlers[dataId] = handler
	return nil
}

// StartWatching 开始监听配置变更
func (w *ConfigWatcher) StartWatching() error {
	if !w.isConnected {
		return fmt.Errorf("nacos client is not connected")
	}

	if len(w.handlers) == 0 {
		return fmt.Errorf("no handlers registered")
	}

	logger.Info("Starting to watch for configuration changes",
		zap.Int("handlersCount", len(w.handlers)),
		zap.String("group", w.config.Group),
		zap.String("namespace", w.config.Namespace))

	// 为每个处理器启动监听
	for dataId, handler := range w.handlers {
		err := w.startWatchingConfig(dataId, handler)
		if err != nil {
			return fmt.Errorf("failed to start watching for %s: %w", dataId, err)
		}
	}

	logger.Info("Successfully started watching for all configuration changes")
	return nil
}

// startWatchingConfig 开始监听特定配置
func (w *ConfigWatcher) startWatchingConfig(dataId string, handler ConfigChangeHandler) error {
	err := w.client.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  w.config.Group,
		OnChange: func(namespace, group, dataId, data string) {
			logger.Info("Configuration change detected",
				zap.String("namespace", namespace),
				zap.String("group", group),
				zap.String("dataId", dataId),
				zap.Int("dataLength", len(data)))

			if err := handler.OnChange(data); err != nil {
				logger.Error("Failed to handle configuration change",
					zap.Error(err),
					zap.String("dataId", dataId))
			}
		},
	})
	if err != nil {
		return fmt.Errorf("failed to listen for config changes: %w", err)
	}

	logger.Info("Successfully started watching for configuration changes",
		zap.String("group", w.config.Group),
		zap.String("dataId", dataId))

	return nil
}

// GetHandler 获取指定数据ID的处理器
func (w *ConfigWatcher) GetHandler(dataId string) (ConfigChangeHandler, bool) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	handler, exists := w.handlers[dataId]
	return handler, exists
}

// GetAllHandlers 获取所有处理器
func (w *ConfigWatcher) GetAllHandlers() map[string]ConfigChangeHandler {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	// 返回副本以避免外部修改
	result := make(map[string]ConfigChangeHandler)
	for k, v := range w.handlers {
		result[k] = v
	}
	return result
}

// IsConnected 返回是否已连接
func (w *ConfigWatcher) IsConnected() bool {
	return w.isConnected
}

// Close 关闭监听器
func (w *ConfigWatcher) Close() error {
	w.isConnected = false
	logger.Info("Config watcher closed")
	return nil
}
