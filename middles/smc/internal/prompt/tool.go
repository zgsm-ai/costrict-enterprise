package prompt

import (
	"time"

	"github.com/zgsm-ai/smc/internal/env"
	"github.com/zgsm-ai/smc/internal/rc"
)

/**
 * Tool configuration structure with metadata and interfaces
 */
type Tool struct {
	Name        string                 `json:"name"`
	Module      string                 `json:"module"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Supports    []string               `json:"supports"`
	Parameters  map[string]interface{} `json:"parameters"`
	Returns     map[string]interface{} `json:"returns"`
	Examples    []string               `json:"examples,omitempty"`
	Restful     *Restful               `json:"restful,omitempty"`
	Grpc        *Grpc                  `json:"grpc,omitempty"`
}

/**
 * RESTful API configuration for tool integration
 */
type Restful struct {
	Url    string `json:"url"`
	Method string `json:"method"`
}

/**
 * gRPC configuration for tool integration
 */
type Grpc struct {
	Url    string `json:"url"`
	Method string `json:"method"`
}

/**
 * Retrieve tool configuration by ID
 * @param toolId unique identifier of the tool
 * @param verbose enable verbose output
 * @return *Tool retrieved tool configuration
 * @return error if retrieval fails
 */
func GetTool(toolId string, verbose bool) (*Tool, error) {
	var p Tool
	err := rc.GetJSON(IDtoKey(toolId, PREFIX_TOOLS), &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

/**
 * List all available tools
 * @return map[string]Tool map of tool IDs to configurations
 * @return error if list operation fails
 */
func ListTools() (map[string]Tool, error) {
	keys, err := rc.KeysByPrefix(PREFIX_TOOLS)
	if err != nil {
		return nil, err
	}
	tools := make(map[string]Tool)
	for _, key := range keys {
		var p Tool
		err := rc.GetJSON(key, &p)
		if err != nil {
			return nil, err
		}
		tools[KeyToID(key, PREFIX_TOOLS)] = p
	}
	return tools, nil
}

/**
 * Register new tool configuration
 * @param toolId unique identifier for the tool
 * @param p tool configuration to register
 * @return error if registration fails
 */
func AddTool(toolId string, p *Tool) error {
	return rc.SetJSON(IDtoKey(toolId, PREFIX_TOOLS), p,
		time.Hour*time.Duration(env.RedisTimeout))
}
