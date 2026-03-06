package prompt

import (
	"time"

	"github.com/zgsm-ai/smc/internal/env"
	"github.com/zgsm-ai/smc/internal/rc"
)

/**
 * Prompt extension configuration with metadata and features
 */
type PromptExtension struct {
	Name          string      `json:"name" description:"Extension name"`
	Publisher     string      `json:"publisher" description:"Publisher name"`
	DisplayName   string      `json:"displayName" description:"Display name of the extension"`
	Icon          string      `json:"icon" description:"Path to extension icon"`
	Description   string      `json:"description" description:"Description of the extension"`
	Version       string      `json:"version" description:"Version number of the extension"`
	ExtensionType string      `json:"extensionType" description:"Type of extension"`
	License       string      `json:"license" description:"License agreement"`
	Engines       Engines     `json:"engines" description:"Engine configuration"`
	Contributes   Contributes `json:"contributes" description:"Extension features"`
}

/**
 * Engine compatibility requirements
 */
type Engines struct {
	Name    string `json:"name" description:"Engine name"`
	Version string `json:"version" description:"Engine version number"`
}

/**
 * Extension feature contributions
 */
type Contributes struct {
	Prompts     []Prompt     `json:"prompts" description:"Prompt templates"`
	Languages   []string     `json:"languages" description:"Supported languages"`
	Dependences []Dependence `json:"dependences" description:"Extension dependencies"`
}

/**
 * Extension dependency configuration
 */
type Dependence struct {
	Name         string `json:"name" description:"Dependency name"`
	Version      string `json:"version" description:"Dependency version number"`
	FailStrategy string `json:"failStrategy" description:"Failure handling strategy"`
}

// Constants for enum values
const (
	ExtensionTypePrompt = "prompt"

	MessageRoleSystem = "system"
	MessageRoleUser   = "user"

	SupportChat       = "chat"
	SupportCodeReview = "codereview"

	FailStrategyAbort  = "abort"
	FailStrategyIgnore = "ignore"
)

/**
 * Retrieve extension by ID
 * @param extensionId unique extension identifier
 * @return *PromptExtension retrieved extension
 * @return error if retrieval fails
 */
func GetExtension(extensionId string) (*PromptExtension, error) {
	var p PromptExtension
	err := rc.GetJSON(IDtoKey(extensionId, PREFIX_EXTENSIONS), &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

/**
 * List all available extensions
 * @return map[string]PromptExtension map of extension IDs to configurations
 * @return error if list operation fails
 */
func ListExtensions() (map[string]PromptExtension, error) {
	keys, err := rc.KeysByPrefix(PREFIX_EXTENSIONS)
	if err != nil {
		return nil, err
	}
	extensions := make(map[string]PromptExtension)
	for _, key := range keys {
		var p PromptExtension
		err := rc.GetJSON(key, &p)
		if err != nil {
			return nil, err
		}
		extensions[KeyToID(key, PREFIX_EXTENSIONS)] = p
	}
	return extensions, nil
}

/**
 * Register new extension
 * @param extensionId unique extension identifier
 * @param p extension configuration
 * @return error if registration fails
 */
func AddExtension(extensionId string, p *PromptExtension) error {
	return rc.SetJSON(IDtoKey(extensionId, PREFIX_EXTENSIONS), p,
		time.Hour*time.Duration(env.RedisTimeout))
}
