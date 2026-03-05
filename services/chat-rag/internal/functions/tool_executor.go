package functions

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/zgsm-ai/chat-rag/internal/client"
	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/model"
)

type ToolExecutor interface {
	DetectTools(ctx context.Context, content string) (bool, string)

	// ExecuteTools executes tools and returns new messages
	ExecuteTools(ctx context.Context, toolName string, content string) (string, error)

	CheckToolReady(ctx context.Context, toolName string) (bool, error)

	GetToolDescription(toolName string) (string, error)

	GetToolCapability(toolName string) (string, error)

	GetToolRule(toolName string) (string, error)

	GetAllTools() []string
}

// GenericToolExecutor Generic tool executor
type GenericToolExecutor struct {
	toolConfig      *config.ToolConfig
	clientFactory   *client.GenericClientFactory
	parameterParser *GenericParameterParser
}

// NewGenericToolExecutor Create new generic tool executor
func NewGenericToolExecutor(toolConfig *config.ToolConfig) *GenericToolExecutor {
	return &GenericToolExecutor{
		toolConfig:      toolConfig,
		clientFactory:   client.NewGenericClientFactory(),
		parameterParser: NewGenericParameterParser(),
	}
}

// DetectTools Detect tool invocation
func (e *GenericToolExecutor) DetectTools(ctx context.Context, content string) (bool, string) {
	for _, toolConfig := range e.toolConfig.GenericTools {
		if strings.Contains(content, "<"+toolConfig.Name+">") {
			return true, toolConfig.Name
		}
	}
	return false, ""
}

// ExecuteTools Execute tools
func (e *GenericToolExecutor) ExecuteTools(ctx context.Context, toolName string, content string) (string, error) {
	// Find tool configuration
	toolConfig, err := e.findToolConfig(toolName)
	if err != nil {
		return "", fmt.Errorf("tool not found: %w", err)
	}

	// Get context parameters
	genericParams, err := e.getGenericParameters(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get context parameters: %w", err)
	}

	// Extract tool parameters, pass context parameters for path parameter processing
	toolParams, err := e.parameterParser.ExtractParametersWithContext(*e.toolConfig, toolName, content, genericParams)
	if err != nil {
		return "", fmt.Errorf("failed to extract parameters: %w", err)
	}

	// Merge parameters
	allParams := make(map[string]interface{})
	for k, v := range toolParams {
		allParams[k] = v
	}
	for k, v := range genericParams {
		allParams[k] = v
	}

	// Validate parameters
	if err := e.parameterParser.ValidateParameters(toolConfig, allParams); err != nil {
		return "", fmt.Errorf("parameter validation failed: %w", err)
	}

	// Get or create client
	toolClient, err := e.clientFactory.CreateClient(toolConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}

	// Execute tool invocation
	result, err := toolClient.Execute(ctx, allParams)
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	return result, nil
}

// CheckToolReady Check tool readiness status
func (e *GenericToolExecutor) CheckToolReady(ctx context.Context, toolName string) (bool, error) {
	// Find tool configuration
	toolConfig, err := e.findToolConfig(toolName)
	if err != nil {
		return false, fmt.Errorf("tool not found: %w", err)
	}

	// Get context parameters
	contextParams, err := e.getGenericParameters(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get context parameters: %w", err)
	}

	// Get or create client
	toolClient, err := e.clientFactory.CreateClient(toolConfig)
	if err != nil {
		return false, fmt.Errorf("failed to create client: %w", err)
	}

	// Check service readiness status
	return toolClient.CheckReady(ctx, contextParams)
}

// GetToolDescription Get tool description
func (e *GenericToolExecutor) GetToolDescription(toolName string) (string, error) {
	toolConfig, err := e.findToolConfig(toolName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("## %s\n%s", toolName, toolConfig.Description), nil
}

// GetToolCapability Get tool capability description
func (e *GenericToolExecutor) GetToolCapability(toolName string) (string, error) {
	toolConfig, err := e.findToolConfig(toolName)
	if err != nil {
		return "", err
	}
	return toolConfig.Capability, nil
}

// GetToolRule Get tool usage rules
func (e *GenericToolExecutor) GetToolRule(toolName string) (string, error) {
	toolConfig, err := e.findToolConfig(toolName)
	if err != nil {
		return "", err
	}
	return toolConfig.Rule, nil
}

// GetAllTools Get all tool names
func (e *GenericToolExecutor) GetAllTools() []string {
	tools := make([]string, 0, len(e.toolConfig.GenericTools))
	for _, config := range e.toolConfig.GenericTools {
		tools = append(tools, config.Name)
	}
	return tools
}

// findToolConfig Find tool configuration
func (e *GenericToolExecutor) findToolConfig(toolName string) (config.GenericToolConfig, error) {
	for _, toolConfig := range e.toolConfig.GenericTools {
		if toolConfig.Name == toolName {
			return toolConfig, nil
		}
	}
	return config.GenericToolConfig{}, fmt.Errorf("tool %s not found", toolName)
}

// getGenericParameters Get context parameters
func (e *GenericToolExecutor) getGenericParameters(ctx context.Context) (map[string]interface{}, error) {
	identity, exists := model.GetIdentityFromContext(ctx)
	if !exists {
		return nil, fmt.Errorf("identity not found in context")
	}

	return map[string]interface{}{
		client.CommonParamClientID:      identity.ClientID,
		client.CommonParamCodebasePath:  identity.ProjectPath,
		client.CommonParamClientVersion: identity.ClientVersion,
		client.CommonParamAuthorization: identity.AuthToken,
	}, nil
}

// GenericParameterParser Generic parameter parser
type GenericParameterParser struct{}

// NewGenericParameterParser Create new parameter parser
func NewGenericParameterParser() *GenericParameterParser {
	return &GenericParameterParser{}
}

// ExtractParameters Extract parameters
func (p *GenericParameterParser) ExtractParameters(toolConfig config.ToolConfig, toolName string, content string) (map[string]interface{}, error) {
	return p.ExtractParametersWithContext(toolConfig, toolName, content, nil)
}

// ExtractParametersWithContext Extract parameters with context information support
func (p *GenericParameterParser) ExtractParametersWithContext(toolConfig config.ToolConfig, toolName string, content string, genericParams map[string]interface{}) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	// Find current tool configuration
	var currentToolConfig *config.GenericToolConfig
	for _, tool := range toolConfig.GenericTools {
		if tool.Name == toolName {
			currentToolConfig = &tool
			break
		}
	}

	if currentToolConfig == nil {
		return nil, fmt.Errorf("tool %s not found in configuration", toolName)
	}

	// Extract XML parameters
	toolContent, err := extractXmlParam(content, currentToolConfig.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tool content: %w", err)
	}

	// Get OS type, default to Windows
	osType := getOSType(genericParams)

	// Extract parameters based on parameter configuration (original specific parameter processing)
	for _, param := range currentToolConfig.Parameters {
		// Handle parameters extracted from LLM
		if param.Source == config.ParameterSourceLLM {
			value, err := extractXmlParam(toolContent, param.Name)
			if err != nil {
				if param.Required {
					return nil, fmt.Errorf("required parameter %s not found: %w", param.Name, err)
				}
				// Optional parameter, use default value
				if param.Default != nil {
					params[param.Name] = param.Default
				}
				continue
			}

			// Special handling for path parameters
			if strings.Contains(strings.ToLower(param.Name), "path") {
				value = p.processPathParameter(value, osType)
			}

			// Type conversion
			convertedValue, err := p.ConvertParameterType(value, param.Type)
			if err != nil {
				return nil, fmt.Errorf("failed to convert parameter %s: %w", param.Name, err)
			}

			params[param.Name] = convertedValue
		} else if param.Source == config.ParameterSourceManual {
			// Handle manually set parameters (get from default field in config file)
			if param.Default != nil {
				params[param.Name] = param.Default
			} else if param.Required {
				return nil, fmt.Errorf("required manual parameter %s must have a default value in configuration", param.Name)
			}
		}
	}

	return params, nil
}

func extractXmlParam(content, paramName string) (string, error) {
	startTag := "<" + paramName + ">"
	endTag := "</" + paramName + ">"

	start := strings.Index(content, startTag)
	if start == -1 {
		return "", fmt.Errorf("start tag not found")
	}

	end := strings.Index(content, endTag)
	if end == -1 {
		return "", fmt.Errorf("end tag not found")
	}

	paramValue := content[start+len(startTag) : end]

	// Check and replace double backslashes with single backslashes to conform to Windows path format
	paramValue = strings.ReplaceAll(paramValue, "\\\\", "\\")

	return paramValue, nil
}

// getOSType Get OS type
func getOSType(contextParams map[string]interface{}) string {
	osType := "windows"
	if contextParams != nil {
		if os, exists := contextParams["osType"]; exists {
			if osStr, ok := os.(string); ok {
				osType = osStr
			}
		}
	}
	return osType
}

// ValidateParameters Validate parameters
func (p *GenericParameterParser) ValidateParameters(toolConfig config.GenericToolConfig, params map[string]interface{}) error {
	for _, param := range toolConfig.Parameters {
		// Check required parameters
		if param.Required {
			if _, exists := params[param.Name]; !exists {
				return fmt.Errorf("required parameter %s is missing", param.Name)
			}
		}

		// Type validation
		if value, exists := params[param.Name]; exists {
			if err := p.validateParameterType(value, param.Type); err != nil {
				return fmt.Errorf("parameter %s type validation failed: %w", param.Name, err)
			}
		}
	}

	return nil
}

// ConvertParameterType Convert parameter type (public method for testing)
func (p *GenericParameterParser) ConvertParameterType(value string, paramType string) (interface{}, error) {
	switch config.ParameterType(strings.ToLower(paramType)) {
	case config.ParameterTypeString:
		return value, nil
	case config.ParameterTypeInteger:
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %s to integer: %w", value, err)
		}
		return intValue, nil
	case config.ParameterTypeFloat:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %s to float: %w", value, err)
		}
		return floatValue, nil
	case config.ParameterTypeBoolean:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %s to boolean: %w", value, err)
		}
		return boolValue, nil
	case config.ParameterTypeArray:
		// Simple array parsing, assume value is comma-separated
		if strings.Contains(value, ",") {
			return strings.Split(value, ","), nil
		}
		return []string{value}, nil
	default:
		return value, nil
	}
}

// validateParameterType Validate parameter type
func (p *GenericParameterParser) validateParameterType(value interface{}, paramType string) error {
	switch config.ParameterType(strings.ToLower(paramType)) {
	case config.ParameterTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case config.ParameterTypeInteger:
		if _, ok := value.(int); !ok {
			return fmt.Errorf("expected integer, got %T", value)
		}
	case config.ParameterTypeFloat:
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("expected float, got %T", value)
		}
	case config.ParameterTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case config.ParameterTypeArray:
		if _, ok := value.([]string); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	}
	return nil
}

// processPathParameter Process special conversion for path parameters
func (p *GenericParameterParser) processPathParameter(path string, osType string) string {
	// If Windows system, convert Unix path separators to Windows path separators
	if strings.Contains(strings.ToLower(osType), "windows") {
		path = strings.ReplaceAll(path, "/", "\\")
	}

	// Handle double backslashes, convert to single backslash
	path = strings.ReplaceAll(path, "\\\\", "\\")

	return path
}
