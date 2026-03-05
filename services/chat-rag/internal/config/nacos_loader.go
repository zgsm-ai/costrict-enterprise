package config

import (
	"fmt"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"go.uber.org/zap"
)

// NacosLoader handles configuration loading from Nacos
type NacosLoader struct {
	client      config_client.IConfigClient
	config      NacosConfig
	handlers    map[string]ConfigChangeHandler
	watcher     *ConfigWatcher
	mutex       sync.RWMutex
	isConnected bool
}

// NewNacosLoader creates a new Nacos configuration loader
func NewNacosLoader(config NacosConfig) (*NacosLoader, error) {
	loader := &NacosLoader{
		config:   config,
		handlers: make(map[string]ConfigChangeHandler),
	}

	// Initialize Nacos client
	err := loader.initClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Nacos client: %w", err)
	}

	// Create config watcher
	loader.watcher = NewConfigWatcher(config, loader.client)

	return loader, nil
}

// initClient initializes the Nacos configuration client
func (nl *NacosLoader) initClient() error {
	// Build server configuration
	serverConfig := []constant.ServerConfig{
		{
			IpAddr:   nl.config.ServerAddr,
			Port:     uint64(nl.config.ServerPort),
			GrpcPort: uint64(nl.config.GrpcPort),
		},
	}
	// Build client configuration
	clientConfig := constant.ClientConfig{
		NamespaceId:         nl.config.Namespace,
		TimeoutMs:           uint64(nl.config.TimeoutSec * 1000),
		NotLoadCacheAtStart: true,
		LogDir:              nl.config.LogDir,
		CacheDir:            nl.config.CacheDir,
		LogLevel:            "debug",
	}

	// Create Nacos config client
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfig,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create Nacos client: %w", err)
	}

	nl.client = client
	nl.isConnected = true

	logger.Info("Nacos client initialized successfully",
		zap.String("serverAddr", nl.config.ServerAddr),
		zap.Int("serverPort", nl.config.ServerPort),
		zap.String("namespace", nl.config.Namespace),
		zap.String("group", nl.config.Group))

	return nil
}

// RegisterConfigHandler registers a configuration change handler
func (nl *NacosLoader) RegisterConfigHandler(handler ConfigChangeHandler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	dataId := handler.GetDataId()
	if dataId == "" {
		return fmt.Errorf("dataId cannot be empty")
	}

	nl.mutex.Lock()
	defer nl.mutex.Unlock()

	if _, exists := nl.handlers[dataId]; exists {
		return fmt.Errorf("handler for dataId %s already registered", dataId)
	}

	// Register to internal map
	nl.handlers[dataId] = handler

	// Register to config watcher
	if err := nl.watcher.RegisterHandler(handler); err != nil {
		delete(nl.handlers, dataId)
		return fmt.Errorf("failed to register handler to watcher: %w", err)
	}

	logger.Info("Configuration handler registered successfully",
		zap.String("dataId", dataId))

	return nil
}

// RegisterGenericConfig registers a generic configuration handler with logging and custom callback
func (nl *NacosLoader) RegisterGenericConfig(dataId string, configType interface{}, onChange func(interface{})) error {
	// Create a generic config handler with default logging using NewGenericConfigHandler
	handler := NewGenericConfigHandler(dataId, configType, func(config interface{}) {
		logger.Info("Configuration updated from Nacos",
			zap.String("dataId", dataId),
			zap.Any("config", config))
		if onChange != nil {
			onChange(config)
		}
	})
	return nl.RegisterConfigHandler(handler)
}

// LoadConfig loads configuration from Nacos
func (nl *NacosLoader) LoadConfig(dataId string, target interface{}) error {
	if !nl.isConnected {
		return fmt.Errorf("nacos client is not connected")
	}

	content, err := nl.client.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  nl.config.Group,
	})
	if err != nil {
		return fmt.Errorf("failed to get %s config from Nacos: %w", dataId, err)
	}

	if content == "" {
		return fmt.Errorf("%s config is empty in Nacos", dataId)
	}

	// Use helper function to parse YAML
	if err := unmarshalYAMLContent(content, target); err != nil {
		return fmt.Errorf("failed to unmarshal %s config: %w", dataId, err)
	}

	logger.Info("Configuration loaded from Nacos successfully",
		zap.String("group", nl.config.Group),
		zap.String("dataId", dataId))

	return nil
}

// StartWatching starts watching for configuration changes (improved version with no parameters)
func (nl *NacosLoader) StartWatching() error {
	if !nl.isConnected {
		return fmt.Errorf("nacos client is not connected")
	}

	if len(nl.handlers) == 0 {
		return fmt.Errorf("no configuration handlers registered")
	}

	logger.Info("Starting to watch for configuration changes",
		zap.Int("handlersCount", len(nl.handlers)),
		zap.String("group", nl.config.Group),
		zap.String("namespace", nl.config.Namespace))

	// Start configuration watching
	if err := nl.watcher.StartWatching(); err != nil {
		return fmt.Errorf("failed to start watching: %w", err)
	}

	logger.Info("Successfully started watching for all configuration changes")
	return nil
}

// Close closes the Nacos client connection
func (nl *NacosLoader) Close() error {
	if nl.client != nil {
		nl.isConnected = false
		logger.Info("Nacos client connection closed")
	}
	return nil
}

// IsConnected returns whether the Nacos client is connected
func (nl *NacosLoader) IsConnected() bool {
	return nl.isConnected
}
