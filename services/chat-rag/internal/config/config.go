package config

import "time"

// ParameterSource Parameter source enumeration
type ParameterSource string

const (
	ParameterSourceLLM    ParameterSource = "llm"    // Extract from LLM response, LLM must provide XML format
	ParameterSourceManual ParameterSource = "manual" // Manual setting, get from default field in config file
)

// ParameterType Parameter type enumeration
type ParameterType string

const (
	ParameterTypeString  ParameterType = "string"
	ParameterTypeInteger ParameterType = "integer"
	ParameterTypeFloat   ParameterType = "float"
	ParameterTypeBoolean ParameterType = "boolean"
	ParameterTypeArray   ParameterType = "array"
)

// LLMConfig
type LLMConfig struct {
	Endpoint            string
	FuncCallingModels   []string
	ChunkMetricsEnabled bool
}

// LLMTimeoutConfig holds idle timeout configuration for LLM requests
type LLMTimeoutConfig struct {
	// Regular mode timeout configuration
	IdleTimeoutMs      int `mapstructure:"idleTimeoutMs" yaml:"idleTimeoutMs"`
	TotalIdleTimeoutMs int `mapstructure:"totalIdleTimeoutMs" yaml:"totalIdleTimeoutMs"`

	// Retry configuration for regular mode
	MaxRetryCount   int `mapstructure:"maxRetryCount" yaml:"maxRetryCount"`
	RetryIntervalMs int `mapstructure:"retryIntervalMs" yaml:"retryIntervalMs"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type ToolConfig struct {
	// Global switch to control whether all tools are disabled, default is false
	DisableTools bool
	// Control which agents in which modes cannot use tools
	DisabledAgents map[string][]string

	// Generic tool configuration
	GenericTools []GenericToolConfig
}

// GenericToolConfig Generic tool configuration structure
type GenericToolConfig struct {
	Name        string                 `yaml:"name"`        // Tool name
	Description string                 `yaml:"description"` // Tool description
	Capability  string                 `yaml:"capability"`  // Tool capability description
	Endpoints   GenericToolEndpoints   `yaml:"endpoints"`   // API endpoint configuration
	Method      string                 `yaml:"method"`      // HTTP request method
	Parameters  []GenericToolParameter `yaml:"parameters"`  // Parameter definitions
	Rule        string                 `yaml:"rule"`        // Tool usage rules
}

// GenericToolEndpoints Tool endpoint configuration
type GenericToolEndpoints struct {
	Search string `yaml:"search"` // Search endpoint
	Ready  string `yaml:"ready"`  // Readiness check endpoint
}

// GenericToolParameter Tool parameter definition
type GenericToolParameter struct {
	Name        string      `yaml:"name"`              // Parameter name
	Type        string      `yaml:"type"`              // Parameter type
	Description string      `yaml:"description"`       // Parameter description
	Required    bool        `yaml:"required"`          // Whether required
	Default     interface{} `yaml:"default,omitempty"` // Default value (optional)
	// Parameter source
	Source ParameterSource `yaml:"source"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	LogFilePath string
	// LogScanIntervalSec   int
	// ClassifyModel        string
	// EnableClassification bool
}

// Deprecated
type ContextCompressConfig struct {
	// Context compression enable flag
	EnableCompress bool
	// Context compression token threshold
	TokenThreshold int
	// Summary Model configuration
	SummaryModel               string
	SummaryModelTokenThreshold int
	// used recent user prompt messages nums
	RecentUserMsgUsedNums int
}

type PreciseContextConfig struct {
	// AgentsMatch configuration
	AgentsMatch []AgentMatchConfig
	// filter "environment_details" user prompt in context
	EnableEnvDetailsFilter bool
	// Control which agents in which modes cannot use ModesChange
	DisabledModesChangeAgents map[string][]string
	// Task content replacement rules
	TaskContentReplaceRule map[string]TaskContentReplaceConfig
}

// TaskContentReplaceConfig holds configuration for task content replacement
type TaskContentReplaceConfig struct {
	// Specify which agents this rule applies to
	ValidAgents map[string][]string `mapstructure:"valid_agents" yaml:"valid_agents"`
	// Skip processing if content contains this key
	SkipKey string `mapstructure:"skip_key" yaml:"skip_key"`
	// Key-value pairs for replacement
	MatchKeys map[string]string `mapstructure:"match_keys" yaml:"match_keys"`
}

// AgentMatchConfig holds configuration for a specific agent matching
type AgentMatchConfig struct {
	Agent string `yaml:"agent"`
	Key   string `yaml:"key"`
}

type FromNacos struct {
	Rules                 *RulesConfig
	Tools                 *ToolConfig
	Router                *RouterConfig
	PreciseContextConfig  *PreciseContextConfig
	VoucherActivityConfig *VoucherActivityConfig
}

// Config holds all service configuration
type Config struct {
	FromNacos

	// Server configuration
	Host string
	Port int

	// Logging configuration
	Log LogConfig

	// Context handling configuration
	ContextCompressConfig ContextCompressConfig

	//Department configuration
	DepartmentApiEndpoint string

	// Redis configuration
	Redis RedisConfig

	LLM LLMConfig

	// LLMTimeout holds idle timeout configuration
	LLMTimeout LLMTimeoutConfig `mapstructure:"llmTimeout" yaml:"llmTimeout"`

	// Forward configuration
	Forward ForwardConfig `mapstructure:"forward" yaml:"forward"`

	// Nacos configuration
	Nacos NacosConfig `mapstructure:"nacos" yaml:"nacos"`
	// Chat metrics reporting configuration
	ChatMetrics ChatMetrics `mapstructure:"chatMetrics" yaml:"chatMetrics"`
	// VIP priority configuration
	VIPPriority VIPPriorityConfig `mapstructure:"vipPriority" yaml:"vipPriority"`

	// Request verification configuration
	RequestVerify RequestVerifyConfig `mapstructure:"requestVerify" yaml:"requestVerify"`
}

// VoucherActivity holds individual voucher activity configuration
type VoucherActivity struct {
	Keyword                string    `mapstructure:"keyword" yaml:"keyword"`                               // Activity keyword (unique identifier)
	CreditAmount           float64   `mapstructure:"creditAmount" yaml:"creditAmount"`                     // Credit amount to redeem
	VoucherExpiryDays      int       `mapstructure:"voucherExpiryDays" yaml:"voucherExpiryDays"`           // Voucher expiry days
	StartTime              time.Time `mapstructure:"startTime" yaml:"startTime"`                           // Activity start time
	EndTime                time.Time `mapstructure:"endTime" yaml:"endTime"`                               // Activity end time
	TotalQuota             int       `mapstructure:"totalQuota" yaml:"totalQuota"`                         // Total redemption quota
	SuccessTemplate        string    `mapstructure:"successTemplate" yaml:"successTemplate"`               // Success response template (Go template format)
	AlreadyRedeemedMessage string    `mapstructure:"alreadyRedeemedMessage" yaml:"alreadyRedeemedMessage"` // Already redeemed message (Go template format)
	ExpiredMessage         string    `mapstructure:"expiredMessage" yaml:"expiredMessage"`                 // Expired message (Go template format)
	QuotaExhaustedMessage  string    `mapstructure:"quotaExhaustedMessage" yaml:"quotaExhaustedMessage"`   // Quota exhausted message (Go template format)
}

// VoucherActivityConfig holds voucher activity configuration
type VoucherActivityConfig struct {
	Enabled    bool              `mapstructure:"enabled" yaml:"enabled"`       // Activity enable flag
	SigningKey string            `mapstructure:"signingKey" yaml:"signingKey"` // Voucher code signing key
	Activities []VoucherActivity `mapstructure:"activities" yaml:"activities"` // List of voucher activities
}

// VoucherRedemptionRecord holds user redemption record for Redis storage
type VoucherRedemptionRecord struct {
	UserID         string    // User ID (UUID)
	UserName       string    // User name
	VoucherCode    string    // Voucher code
	RedemptionTime time.Time // Redemption time
}

// ActivityStatistics holds activity statistics for Redis storage
type ActivityStatistics struct {
	TotalRedeemed  int // Total redeemed count
	RemainingQuota int // Remaining quota
}

// RouterConfig holds router related configuration
type RouterConfig struct {
	Enabled  bool           `mapstructure:"enabled" yaml:"enabled"`
	Strategy string         `mapstructure:"strategy" yaml:"strategy"`
	Semantic SemanticConfig `mapstructure:"semantic" yaml:"semantic"`
	Priority PriorityConfig `mapstructure:"priority" yaml:"priority"`
}

// SemanticConfig holds semantic router strategy configuration
type SemanticConfig struct {
	Analyzer        AnalyzerConfig        `mapstructure:"analyzer" yaml:"analyzer"`
	InputExtraction InputExtractionConfig `mapstructure:"inputExtraction" yaml:"inputExtraction"`
	Routing         RoutingConfig         `mapstructure:"routing" yaml:"routing"`
	RuleEngine      RuleEngineConfig      `mapstructure:"ruleEngine" yaml:"ruleEngine"`
}

// AnalyzerConfig only keeps model and timeoutMs per requirements
type AnalyzerConfig struct {
	Model          string `mapstructure:"model" yaml:"model"`
	TimeoutMs      int    `mapstructure:"timeoutMs" yaml:"timeoutMs"`
	TotalTimeoutMs int    `mapstructure:"totalTimeoutMs" yaml:"totalTimeoutMs"`
	MaxInputBytes  int    `mapstructure:"maxInputBytes" yaml:"maxInputBytes"`
	// Optional: override endpoint and token for analyzer-only requests
	Endpoint string `mapstructure:"endpoint" yaml:"endpoint"`
	ApiToken string `mapstructure:"apiToken" yaml:"apiToken"`
	// Optional fields; ignored if empty
	PromptTemplate string               `mapstructure:"promptTemplate" yaml:"promptTemplate"`
	AnalysisLabels []string             `mapstructure:"analysisLabels" yaml:"analysisLabels"`
	DynamicMetrics DynamicMetricsConfig `mapstructure:"dynamicMetrics" yaml:"dynamicMetrics"`
}

// InputExtractionConfig controls how to extract input and history
type InputExtractionConfig struct {
	Protocol        string `mapstructure:"protocol" yaml:"protocol"`
	UserJoinSep     string `mapstructure:"userJoinSep" yaml:"userJoinSep"`
	StripCodeFences bool   `mapstructure:"stripCodeFences" yaml:"stripCodeFences"`
	CodeFenceRegex  string `mapstructure:"codeFenceRegex" yaml:"codeFenceRegex"`
	MaxUserMessages int    `mapstructure:"maxUserMessages" yaml:"maxUserMessages"`
	MaxHistoryBytes int    `mapstructure:"maxHistoryBytes" yaml:"maxHistoryBytes"`
	// MaxHistoryMessages limits how many history entries (after processing) can be included.
	// When >0, only the most recent N history items are kept.
	MaxHistoryMessages int `mapstructure:"maxHistoryMessages" yaml:"maxHistoryMessages"`
}

// RoutingConfig holds candidate model routing configuration
type RoutingConfig struct {
	Candidates        []RoutingCandidate `mapstructure:"candidates" yaml:"candidates"`
	MinScore          int                `mapstructure:"minScore" yaml:"minScore"`
	TieBreakOrder     []string           `mapstructure:"tieBreakOrder" yaml:"tieBreakOrder"`
	FallbackModelName string             `mapstructure:"fallbackModelName" yaml:"fallbackModelName"`

	// Timeout configuration for model degradation scenarios
	IdleTimeoutMs      int `mapstructure:"idleTimeoutMs" yaml:"idleTimeoutMs"`
	TotalIdleTimeoutMs int `mapstructure:"totalIdleTimeoutMs" yaml:"totalIdleTimeoutMs"`

	// Retry configuration for model degradation scenarios
	MaxRetryCount   int `mapstructure:"maxRetryCount" yaml:"maxRetryCount"`
	RetryIntervalMs int `mapstructure:"retryIntervalMs" yaml:"retryIntervalMs"`
}

// RoutingCandidate defines a candidate model and its scores
type RoutingCandidate struct {
	ModelName string         `mapstructure:"modelName" yaml:"modelName"`
	Enabled   bool           `mapstructure:"enabled" yaml:"enabled"`
	Scores    map[string]int `mapstructure:"scores" yaml:"scores"`
}

// RuleEngineConfig is optional and configurable
type RuleEngineConfig struct {
	Enabled      bool     `mapstructure:"enabled" yaml:"enabled"`
	InlineRules  []string `mapstructure:"inlineRules" yaml:"inlineRules"`
	BodyPrefix   string   `mapstructure:"bodyPrefix" yaml:"bodyPrefix"`
	HeaderPrefix string   `mapstructure:"headerPrefix" yaml:"headerPrefix"`
}

// DynamicMetricsConfig controls dynamic metrics loading for candidate filtering
type DynamicMetricsConfig struct {
	Enabled     bool     `mapstructure:"enabled" yaml:"enabled"`
	RedisPrefix string   `mapstructure:"redisPrefix" yaml:"redisPrefix"`
	Metrics     []string `mapstructure:"metrics" yaml:"metrics"`
}

// PriorityConfig holds priority router strategy configuration
type PriorityConfig struct {
	Candidates         []PriorityCandidate `mapstructure:"candidates" yaml:"candidates"`
	FallbackModelName  string              `mapstructure:"fallbackModelName" yaml:"fallbackModelName"`
	IdleTimeoutMs      int                 `mapstructure:"idleTimeoutMs" yaml:"idleTimeoutMs"`
	TotalIdleTimeoutMs int                 `mapstructure:"totalIdleTimeoutMs" yaml:"totalIdleTimeoutMs"`
	MaxRetryCount      int                 `mapstructure:"maxRetryCount" yaml:"maxRetryCount"`
	RetryIntervalMs    int                 `mapstructure:"retryIntervalMs" yaml:"retryIntervalMs"`
}

// PriorityCandidate defines a candidate model with priority and weight
type PriorityCandidate struct {
	ModelName string `mapstructure:"modelName" yaml:"modelName"`
	Enabled   bool   `mapstructure:"enabled" yaml:"enabled"`
	Priority  int    `mapstructure:"priority" yaml:"priority"` // Lower number = higher priority
	Weight    int    `mapstructure:"weight" yaml:"weight"`     // Weight for round-robin within same priority
}

// AgentConfig holds configuration for a specific agent
type AgentConfig struct {
	MatchAgents []string `mapstructure:"match_agents"`
	MatchModes  []string `mapstructure:"match_modes"`
	Rules       string   `mapstructure:"rules"`
}

// RulesConfig holds the rules configuration for agents
type RulesConfig struct {
	Agents []AgentConfig `yaml:"agents"`
}

// ForwardConfig holds forwarding configuration
type ForwardConfig struct {
	DefaultTarget string `yaml:"defaultTarget"`
	Enabled       bool   `yaml:"enabled"`
}

// NacosConfig holds Nacos configuration center connection settings
type NacosConfig struct {
	// Nacos server address
	ServerAddr string `mapstructure:"serverAddr" yaml:"serverAddr"`
	// Nacos server port
	ServerPort int `mapstructure:"serverPort" yaml:"serverPort"`
	// Nacos gRPC port
	GrpcPort int `mapstructure:"grpcPort" yaml:"grpcPort"`
	// Nacos namespace
	Namespace string `mapstructure:"namespace" yaml:"namespace"`
	// Nacos group
	Group string `mapstructure:"group" yaml:"group"`
	// Timeout in seconds for Nacos operations
	TimeoutSec int `mapstructure:"timeoutSec" yaml:"timeoutSec"`
	// Log directory for Nacos client
	LogDir string `mapstructure:"logDir" yaml:"logDir"`
	// Cache directory for Nacos client
	CacheDir string `mapstructure:"cacheDir" yaml:"cacheDir"`
}

type ChatMetrics struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	Url     string `mapstructure:"url" yaml:"url"`
	Method  string `mapstructure:"method" yaml:"method"`
}

// VIPPriorityConfig holds VIP priority configuration
type VIPPriorityConfig struct {
	Enabled bool `yaml:"enabled"` // Enable setting priority for VIP users
}

type RequestVerifyConfig struct {
	Enabled           bool `yaml:"enabled"`           // Enable request verification
	EnabledTimeVerify bool `yaml:"enabledTimeVerify"` // Enable timestamp verification
}
