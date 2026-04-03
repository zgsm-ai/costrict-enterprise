package bootstrap

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zgsm-ai/chat-rag/internal/client"
	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/functions"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/service"
	"github.com/zgsm-ai/chat-rag/internal/tokenizer"
	"go.uber.org/zap"
)

// ServiceContext holds all service dependencies with thread-safe access
// Fields are exported for backward compatibility while maintaining thread safety through update methods
type ServiceContext struct {
	Config config.Config

	// Clients
	RedisClient client.RedisInterface

	// Services
	LoggerService  service.LogRecordInterface
	MetricsService service.MetricsInterface
	VoucherService *service.VoucherService

	// Utilities
	TokenCounter *tokenizer.TokenCounter

	ToolExecutor functions.ToolExecutor

	// Router strategy instance (maintained as singleton for state consistency)
	// This ensures round-robin and other stateful strategies maintain their state across requests
	// Stored as interface{} to avoid circular dependency with router package
	routerStrategy     interface{}
	routerStrategyLock sync.RWMutex

	// Unified Nacos configuration manager
	NacosConfigManager *NacosConfigManager

	// Lifecycle management (internal fields)
	mu        sync.RWMutex
	isRunning bool
	stopOnce  sync.Once
	stopChan  chan struct{}

	// Error aggregation for initialization
	initErrors []error
}

// ServiceContextOption defines functional options for ServiceContext
type ServiceContextOption func(*ServiceContext) error

// NewServiceContext creates a new service context with all dependencies using builder pattern
// Any initialization failure will panic to prevent service startup with invalid configuration
func NewServiceContext(c config.Config, opts ...ServiceContextOption) *ServiceContext {
	svc := &ServiceContext{
		Config:     c,
		stopChan:   make(chan struct{}),
		initErrors: make([]error, 0),
	}

	// Apply functional options - panic on any option failure
	for _, opt := range opts {
		if err := opt(svc); err != nil {
			panic(fmt.Sprintf("Failed to apply service context option: %v", err))
		}
	}

	// Initialize all components - panic on any initialization failure
	if err := svc.initialize(); err != nil {
		// Cleanup any partially initialized components before panic
		svc.cleanupOnError()
		panic(fmt.Sprintf("Failed to initialize service context: %v", err))
	}

	svc.isRunning = true
	return svc
}

// initialize orchestrates the initialization of all components
func (svc *ServiceContext) initialize() error {
	logger.Info("Starting service context initialization")

	// Initialize components in dependency order
	initializers := []func() error{
		svc.initializeTokenCounter,
		svc.initializeMetricsService,
		svc.initializeLoggerService,
		svc.initializeRedisClient,
		svc.initializeNacosConfig,
		svc.initializeVoucherService,
		svc.initializeToolExecutor,
		svc.initializeRouterStrategy,
		svc.startNacosConfigWatching,
	}

	for _, init := range initializers {
		if err := init(); err != nil {
			svc.initErrors = append(svc.initErrors, err)
			return err
		}
	}

	logger.Info("Service context initialization completed successfully")
	return nil
}

// initializeTokenCounter initializes the token counter with fallback
func (svc *ServiceContext) initializeTokenCounter() error {
	if svc.TokenCounter != nil {
		return nil // Already set via option
	}

	counter, err := tokenizer.NewTokenCounter()
	if err != nil {
		logger.Error("Failed to create token counter, using fallback",
			zap.Error(err))
		// In production, you might want to use a fallback implementation
		// For now, we'll return the error
		return fmt.Errorf("failed to initialize token counter: %w", err)
	}

	svc.TokenCounter = counter
	logger.Info("Token counter initialized successfully")
	return nil
}

// initializeMetricsService initializes the metrics service
func (svc *ServiceContext) initializeMetricsService() error {
	svc.MetricsService = service.NewMetricsService()
	logger.Info("Metrics service initialized successfully")
	return nil
}

// initializeLoggerService initializes and starts the logger service
func (svc *ServiceContext) initializeLoggerService() error {
	svc.LoggerService = service.NewLogRecordService(svc.Config)
	svc.LoggerService.SetMetricsService(svc.MetricsService)

	if err := svc.LoggerService.Start(); err != nil {
		return fmt.Errorf("failed to start logger service: %w", err)
	}

	logger.Info("Logger service initialized and started successfully")
	return nil
}

// initializeRedisClient initializes the Redis client
func (svc *ServiceContext) initializeRedisClient() error {
	if svc.RedisClient != nil {
		return nil // Already set via option
	}

	svc.RedisClient = client.NewRedisClient(svc.Config.Redis)
	logger.Info("Redis client initialized successfully",
		zap.String("addr", svc.Config.Redis.Addr))
	return nil
}

// initializeNacosConfig initializes Nacos configuration
func (svc *ServiceContext) initializeNacosConfig() error {
	// Check if Nacos is configured
	if svc.Config.Nacos.ServerAddr == "" || svc.Config.Nacos.ServerPort <= 0 {
		return fmt.Errorf("nacos is not configured, serverAddr or serverPort is empty")
	}

	// Create unified Nacos config manager
	nacosManager, err := NewNacosConfigManager(svc.Config)
	if err != nil {
		return fmt.Errorf("failed to create Nacos config manager: %w", err)
	}
	svc.NacosConfigManager = nacosManager

	// Initialize configurations
	nacosResult, err := svc.NacosConfigManager.InitializeNacosConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize Nacos configurations: %w", err)
	}

	// Assign configuration results with pointer checks
	if nacosResult.RulesConfig == nil {
		return fmt.Errorf("nacos rules configuration is nil")
	}
	if nacosResult.ToolsConfig == nil {
		return fmt.Errorf("nacos tools configuration is nil")
	}
	if nacosResult.PreciseContextConfig == nil {
		return fmt.Errorf("nacos precise context configuration is nil")
	}
	if nacosResult.RouterConfig == nil {
		return fmt.Errorf("nacos router configuration is nil")
	}
	if nacosResult.VoucherActivityConfig == nil {
		return fmt.Errorf("nacos voucher activity configuration is nil")
	}

	svc.Config.Rules = nacosResult.RulesConfig
	svc.Config.Tools = nacosResult.ToolsConfig
	svc.Config.PreciseContextConfig = nacosResult.PreciseContextConfig
	svc.Config.Router = nacosResult.RouterConfig
	svc.Config.VoucherActivityConfig = nacosResult.VoucherActivityConfig

	// Apply router defaults after loading from Nacos
	config.ApplyRouterDefaults(&svc.Config)

	logger.Info("Nacos configuration initialized successfully",
		zap.String("serverAddr", svc.Config.Nacos.ServerAddr),
		zap.Int("serverPort", svc.Config.Nacos.ServerPort),
		zap.Any("nacosResult", nacosResult))

	return nil
}

// initializeToolExecutor initializes the tool executor
func (svc *ServiceContext) initializeToolExecutor() error {
	svc.ToolExecutor = functions.NewGenericToolExecutor(svc.Config.Tools)
	logger.Info("Tool executor initialized successfully")
	return nil
}

// initializeRouterStrategy initializes the router strategy placeholder
// The actual strategy will be set by the router package on first use
func (svc *ServiceContext) initializeRouterStrategy() error {
	// Strategy will be initialized lazily by router package to avoid circular dependency
	logger.Info("Router strategy initialization deferred to first use")
	return nil
}

// initializeVoucherService initializes the voucher service
func (svc *ServiceContext) initializeVoucherService() error {
	// Get signing key from VoucherActivityConfig
	signingKey := svc.Config.VoucherActivityConfig.SigningKey
	if signingKey == "" {
		return fmt.Errorf("VoucherActivityConfig.SigningKey is not configured")
	}

	// Create voucher service
	svc.VoucherService = service.NewVoucherService(signingKey)
	logger.Info("Voucher service initialized successfully")
	return nil
}

// GetRouterStrategy returns the router strategy instance (thread-safe)
// Returns interface{} to avoid circular dependency - caller should type assert
func (svc *ServiceContext) GetRouterStrategy() interface{} {
	svc.routerStrategyLock.RLock()
	defer svc.routerStrategyLock.RUnlock()
	return svc.routerStrategy
}

// SetRouterStrategy updates the router strategy instance (thread-safe)
// This is called by the router package on first use or when configuration is updated
func (svc *ServiceContext) SetRouterStrategy(strategy interface{}) {
	svc.routerStrategyLock.Lock()
	defer svc.routerStrategyLock.Unlock()
	svc.routerStrategy = strategy
	logger.Info("Router strategy updated")
}

// startNacosConfigWatching starts watching for Nacos configuration changes
func (svc *ServiceContext) startNacosConfigWatching() error {
	if svc.NacosConfigManager == nil {
		return fmt.Errorf("nacos config manager is not available, cannot start configuration watching")
	}

	if err := svc.NacosConfigManager.StartWatching(svc); err != nil {
		return fmt.Errorf("failed to start watching for configuration changes: %w", err)
	}

	logger.Info("Nacos configuration watching started successfully")
	return nil
}

// cleanupOnError cleans up any partially initialized components
func (svc *ServiceContext) cleanupOnError() {
	logger.Info("Cleaning up partially initialized components due to initialization error")

	// Stop logger service if it was started
	if svc.LoggerService != nil {
		svc.LoggerService.Stop()
	}

	// Close Redis connection if it was created
	if svc.RedisClient != nil {
		if err := svc.RedisClient.Close(); err != nil {
			logger.Error("Failed to close Redis connection during cleanup",
				zap.Error(err))
		}
	}

	// Close Nacos connection if it was created
	if svc.NacosConfigManager != nil {
		if err := svc.NacosConfigManager.Stop(); err != nil {
			logger.Error("Failed to close Nacos connection during cleanup",
				zap.Error(err))
		}
	}
}

// Stop gracefully stops all services with proper ordering and timeout
func (svc *ServiceContext) Stop() error {
	var stopError error

	svc.stopOnce.Do(func() {
		logger.Info("Starting graceful shutdown of all services...")
		close(svc.stopChan)

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Define shutdown sequence with proper ordering
		shutdownSteps := []struct {
			name string
			fn   func(context.Context) error
		}{
			{"logger service", svc.shutdownLoggerService},
			{"Nacos connection", svc.shutdownNacosConnection},
			{"Redis connection", svc.shutdownRedisConnection},
		}

		// Execute shutdown steps
		for _, step := range shutdownSteps {
			select {
			case <-ctx.Done():
				stopError = fmt.Errorf("shutdown timeout reached during %s", step.name)
				logger.Error("Shutdown timeout reached",
					zap.String("step", step.name))
				return
			default:
				if err := step.fn(ctx); err != nil {
					logger.Error("Failed to shutdown component",
						zap.String("component", step.name),
						zap.Error(err))
					// Continue with other components even if one fails
				}
			}
		}

		svc.isRunning = false
		logger.Info("All services have been gracefully stopped")
	})

	return stopError
}

// shutdownLoggerService stops the logger service
func (svc *ServiceContext) shutdownLoggerService(ctx context.Context) error {
	if svc.LoggerService == nil {
		return nil
	}

	logger.Info("Stopping logger service...")
	svc.LoggerService.Stop()
	logger.Info("Logger service stopped")
	return nil
}

// shutdownNacosConnection closes the Nacos connection
func (svc *ServiceContext) shutdownNacosConnection(ctx context.Context) error {
	if svc.NacosConfigManager == nil {
		return nil
	}

	logger.Info("Closing Nacos connection...")
	if err := svc.NacosConfigManager.Stop(); err != nil {
		logger.Error("Failed to close Nacos connection",
			zap.Error(err))
		return err
	}

	logger.Info("Nacos connection closed successfully")
	return nil
}

// shutdownRedisConnection closes the Redis connection
func (svc *ServiceContext) shutdownRedisConnection(ctx context.Context) error {
	if svc.RedisClient == nil {
		return nil
	}

	logger.Info("Closing Redis connections...")
	if err := svc.RedisClient.Close(); err != nil {
		logger.Error("Failed to close Redis connection",
			zap.Error(err))
		return err
	}

	logger.Info("Redis connections closed successfully")
	return nil
}

// Direct field access for backward compatibility
// These fields can be accessed directly while maintaining thread safety through the update methods

// Safe accessors for components with thread safety (only include methods that are actually used)
func (svc *ServiceContext) GetConfig() config.Config {
	svc.mu.RLock()
	defer svc.mu.RUnlock()
	return svc.Config
}

// Update methods for Nacos configuration changes (with thread safety)
func (svc *ServiceContext) updateRulesConfig(config *config.RulesConfig) {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	svc.Config.Rules = config
}

func (svc *ServiceContext) updateToolExecutor(executor functions.ToolExecutor) {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	svc.ToolExecutor = executor
}

func (svc *ServiceContext) updatePreciseContextConfig(config *config.PreciseContextConfig) {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	svc.Config.PreciseContextConfig = config
}

func (svc *ServiceContext) updateRouterConfig(routerConfig *config.RouterConfig) {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	svc.Config.Router = routerConfig

	// Apply router defaults after updating from Nacos
	config.ApplyRouterDefaults(&svc.Config)

	// Clear cached router strategy so it will be recreated on next use with new config
	svc.SetRouterStrategy(nil)
	logger.Info("Router configuration updated, strategy cache cleared")
}

func (svc *ServiceContext) updateVoucherActivityConfig(newConfig *config.VoucherActivityConfig) {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	svc.Config.VoucherActivityConfig = newConfig
	svc.initializeVoucherService()
	logger.Info("Voucher activity configuration updated")
}
