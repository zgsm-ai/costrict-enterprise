package bootstrap

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/functions"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"go.uber.org/zap"
)

// NacosConfigResult holds the result of Nacos configuration initialization
type NacosConfigResult struct {
	RulesConfig           *config.RulesConfig
	ToolsConfig           *config.ToolConfig
	PreciseContextConfig  *config.PreciseContextConfig
	RouterConfig          *config.RouterConfig
	VoucherActivityConfig *config.VoucherActivityConfig
}

// NacosConfigMetadata holds metadata for Nacos configuration registration
type NacosConfigMetadata struct {
	DataId     string
	ConfigType interface{}
	UpdateFunc func(svc *ServiceContext, config interface{})
}

// NacosConfigManager handles all Nacos configuration management operations
// This is a unified manager that encapsulates both loader and configuration management
type NacosConfigManager struct {
	nacosLoader *config.NacosLoader
	config      config.Config
	stopChan    chan struct{}
	stopOnce    sync.Once
}

// NewNacosConfigManager creates a new Nacos configuration manager
func NewNacosConfigManager(cfg config.Config) (*NacosConfigManager, error) {
	manager := &NacosConfigManager{
		config:   cfg,
		stopChan: make(chan struct{}),
	}

	// Check if Nacos is configured
	if !manager.isNacosConfigured() {
		return nil, fmt.Errorf("nacos is not configured, serverAddr or serverPort is empty")
	}

	// Create Nacos loader using the package-level function
	loader, err := config.NewNacosLoader(cfg.Nacos)
	if err != nil {
		return nil, fmt.Errorf("failed to create Nacos loader: %w", err)
	}

	manager.nacosLoader = loader

	logger.Info("Nacos configuration manager created successfully",
		zap.String("serverAddr", cfg.Nacos.ServerAddr),
		zap.Int("serverPort", cfg.Nacos.ServerPort))

	return manager, nil
}

// isNacosConfigured checks if Nacos is properly configured
func (m *NacosConfigManager) isNacosConfigured() bool {
	return m.config.Nacos.ServerAddr != "" && m.config.Nacos.ServerPort > 0
}

// InitializeNacosConfig loads all configurations from Nacos
func (m *NacosConfigManager) InitializeNacosConfig() (*NacosConfigResult, error) {
	logger.Info("Initializing Nacos configurations")

	metadataList := getNacosConfigMetadata()
	result, err := m.loadAllConfigurations(metadataList)
	if err != nil {
		panic(fmt.Sprintf("Failed to load Nacos configurations: %v", err))
	}

	logger.Info("Nacos configuration initialization completed successfully")
	return result, nil
}

// StartWatching starts watching for configuration changes
func (m *NacosConfigManager) StartWatching(svc *ServiceContext) error {
	metadataList := getNacosConfigMetadata()

	// Register all configurations
	if err := m.registerAllConfigurations(metadataList, svc); err != nil {
		return fmt.Errorf("failed to register configurations: %w", err)
	}

	// Start watching
	if err := m.nacosLoader.StartWatching(); err != nil {
		return fmt.Errorf("failed to start watching for configuration changes: %w", err)
	}

	logger.Info("Nacos configuration watching started successfully")
	return nil
}

// Stop gracefully stops the Nacos configuration manager
func (m *NacosConfigManager) Stop() error {
	var err error

	m.stopOnce.Do(func() {
		close(m.stopChan)

		if m.nacosLoader != nil {
			logger.Info("Closing Nacos connection...")
			if closeErr := m.nacosLoader.Close(); closeErr != nil {
				// 配置管理器关闭失败也应该panic，因为这会影响整个系统
				panic(fmt.Sprintf("Failed to close Nacos connection: %v", closeErr))
			} else {
				logger.Info("Nacos connection closed successfully")
			}
		}

	})

	return err
}

// loadAllConfigurations loads all configurations from Nacos using metadata
func (m *NacosConfigManager) loadAllConfigurations(metadataList []NacosConfigMetadata) (*NacosConfigResult, error) {
	result := &NacosConfigResult{}

	// Load all configurations using metadata with reflection
	for _, metadata := range metadataList {
		// Create a new instance of the config type
		configInstance := metadata.ConfigType
		if err := m.nacosLoader.LoadConfig(metadata.DataId, configInstance); err != nil {
			return nil, fmt.Errorf("failed to load %s from Nacos: %w", metadata.DataId, err)
		}

		// Use reflection to automatically assign to result fields based on type
		assignConfigToResult(result, configInstance)
	}

	return result, nil
}

// registerAllConfigurations registers all configuration watchers using metadata
func (m *NacosConfigManager) registerAllConfigurations(metadataList []NacosConfigMetadata, svc *ServiceContext) error {
	// Register all configurations using the factory method
	for _, metadata := range metadataList {
		if err := m.registerConfig(metadata, svc); err != nil {
			return fmt.Errorf("failed to register configuration %s: %w", metadata.DataId, err)
		}
	}
	return nil
}

// registerConfig registers a single Nacos configuration
func (m *NacosConfigManager) registerConfig(metadata NacosConfigMetadata, svc *ServiceContext) error {
	return m.nacosLoader.RegisterGenericConfig(
		metadata.DataId,
		metadata.ConfigType,
		func(data interface{}) {
			metadata.UpdateFunc(svc, data)
			logger.Info(fmt.Sprintf("Configuration %s updated successfully", metadata.DataId),
				zap.String("dataId", metadata.DataId))
		},
	)
}

// getNacosConfigMetadata returns the centralized configuration metadata
// This is the only place that needs to be modified when adding new configurations
// All DataId strings are hardcoded here - when adding new configurations, only modify this function
func getNacosConfigMetadata() []NacosConfigMetadata {
	return []NacosConfigMetadata{
		{
			DataId:     "agent_rules",
			ConfigType: &config.RulesConfig{},
			UpdateFunc: func(svc *ServiceContext, data interface{}) {
				if rulesConfig, ok := data.(*config.RulesConfig); ok {
					svc.updateRulesConfig(rulesConfig)
					logger.Info("Agent rules configuration updated",
						zap.Int("agentsCount", len(rulesConfig.Agents)))
				}
			},
		},
		{
			DataId:     "tools_prompt",
			ConfigType: &config.ToolConfig{},
			UpdateFunc: func(svc *ServiceContext, data interface{}) {
				if toolsConfig, ok := data.(*config.ToolConfig); ok {
					logger.Info("Recreating tool executor with new tools configuration")
					newToolExecutor := functions.NewGenericToolExecutor(toolsConfig)
					svc.updateToolExecutor(newToolExecutor)
					logger.Info("Tool executor successfully recreated with new configuration")
				}
			},
		},
		{
			DataId:     "precise_context",
			ConfigType: &config.PreciseContextConfig{},
			UpdateFunc: func(svc *ServiceContext, data interface{}) {
				if preciseContextConfig, ok := data.(*config.PreciseContextConfig); ok {
					svc.updatePreciseContextConfig(preciseContextConfig)
					logger.Info("Precise context configuration updated",
						zap.Int("agentsMatchCount", len(preciseContextConfig.AgentsMatch)),
						zap.Bool("envDetailsFilterEnabled", preciseContextConfig.EnableEnvDetailsFilter))
				}
			},
		},
		{
			DataId:     "model_router",
			ConfigType: &config.RouterConfig{},
			UpdateFunc: func(svc *ServiceContext, data interface{}) {
				if routerConfig, ok := data.(*config.RouterConfig); ok {
					svc.updateRouterConfig(routerConfig)
					logger.Info("Router configuration updated",
						zap.Bool("enabled", routerConfig.Enabled),
						zap.String("strategy", routerConfig.Strategy))
				}
			},
		},
		{
			DataId:     "voucher_activity",
			ConfigType: &config.VoucherActivityConfig{},
			UpdateFunc: func(svc *ServiceContext, data interface{}) {
				if voucherActivityConfig, ok := data.(*config.VoucherActivityConfig); ok {
					svc.updateVoucherActivityConfig(voucherActivityConfig)
					logger.Info("Voucher activity configuration updated",
						zap.Bool("enabled", voucherActivityConfig.Enabled),
						zap.Int("activities_count", len(voucherActivityConfig.Activities)))
				}
			},
		},
	}
}

// assignConfigToResult uses reflection to automatically assign config instances to result fields
// This function eliminates the need for manual switch cases when adding new configuration types
func assignConfigToResult(result *NacosConfigResult, configInstance interface{}) {
	// Use reflection to get the type and value of the config instance
	configValue := reflect.ValueOf(configInstance)
	if configValue.Kind() != reflect.Ptr || configValue.IsNil() {
		return
	}

	// Get the config instance type (pointer type)
	configType := configValue.Type()

	// Use reflection to set the corresponding field in result
	resultValue := reflect.ValueOf(result).Elem()
	resultType := resultValue.Type()

	// Find the matching field in NacosConfigResult and assign the config
	for i := 0; i < resultType.NumField(); i++ {
		field := resultType.Field(i)
		fieldValue := resultValue.Field(i)

		// Check if the field type matches our config type (both should be pointer types)
		if field.Type == configType && fieldValue.CanSet() {
			fieldValue.Set(configValue)
			logger.Info("Configuration assigned to result field",
				zap.String("field", field.Name),
				zap.String("type", configType.String()))
			return
		}
	}

	logger.Warn("No matching field found for configuration type",
		zap.String("type", configType.String()))
}
