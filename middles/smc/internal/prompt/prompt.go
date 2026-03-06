package prompt

import (
	"time"

	"github.com/zgsm-ai/smc/internal/env"
	"github.com/zgsm-ai/smc/internal/rc"
)

/**
 * Chat message structure with role and content
 */
type Message struct {
	Role    string `json:"role" description:"message role (system/user/assistant)"`
	Content string `json:"content" description:"message content text"`
}

/**
 * Prompt template structure including configuration and parameters
 */
type Prompt struct {
	Name       string                 `json:"name" description:"prompt template name"`
	Messages   []Message              `json:"messages,omitempty" description:"messages list"`
	Prompt     string                 `json:"prompt,omitempty" description:"user prompt template"`
	Supports   []string               `json:"supports" description:"supported scenarios"`
	Parameters map[string]interface{} `json:"parameters" description:"parameter definition (JSON Schema)"`
	Returns    map[string]interface{} `json:"returns" description:"return value definition (JSON Schema)"`
}

/**
 * Retrieve prompt template by ID
 * @param promptId unique identifier for prompt
 * @param verbose enable verbose output
 * @return *Prompt retrieved prompt template
 * @return error if retrieval fails
 */
func GetPrompt(promptId string, verbose bool) (*Prompt, error) {
	var p Prompt
	err := rc.GetJSON(IDtoKey(promptId, PREFIX_TEMPLATES), &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

/**
 * List all available prompt templates
 * @return map[string]Prompt map of prompt IDs to templates
 * @return error if list operation fails
 */
func ListPrompts() (map[string]Prompt, error) {
	keys, err := rc.KeysByPrefix(PREFIX_TEMPLATES)
	if err != nil {
		return nil, err
	}
	prompts := make(map[string]Prompt)
	for _, key := range keys {
		var p Prompt
		err := rc.GetJSON(key, &p)
		if err != nil {
			return nil, err
		}
		prompts[KeyToID(key, PREFIX_TEMPLATES)] = p
	}
	return prompts, nil
}

/**
 * Add new prompt template to storage
 * @param promptId unique identifier for new prompt
 * @param p prompt template to add
 * @return error if add operation fails
 */
func AddPrompt(promptId string, p *Prompt) error {
	return rc.SetJSON(IDtoKey(promptId, PREFIX_TEMPLATES), p,
		time.Hour*time.Duration(env.RedisTimeout))
}
