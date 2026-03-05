package config

import (
	"bytes"
	"completion-agent/pkg/logger"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

var CostrictDir string = getCostrictDir()

var globalEnv map[string]interface{} = map[string]interface{}{
	"Env": map[string]interface{}{
		"CostrictDir": getCostrictDir(),
		"CodebaseUrl": getCodebaseUrl(),
	},
	"Auth": map[string]interface{}{
		"BaseUrl":     GetBaseURL(),
		"AccessToken": GetAuthConfig().AccessToken,
		"ID":          GetAuthConfig().ID,
		"Name":        GetAuthConfig().Name,
		"MachineID":   GetAuthConfig().MachineID,
	},
}

/**
 * Load system knowledge from well-known configuration file
 * @returns {SystemKnowledge, error} Returns system knowledge structure and error if any
 * @description
 * - Loads knowledge from .costrict/share/.well-known.json file
 * - Parses JSON content into SystemKnowledge structure
 * - Returns error if file doesn't exist or JSON parsing fails
 * @example
 * knowledge, err := loadKnowledge()
 * if err != nil {
 *     log.Printf("Failed to load knowledge: %v", err)
 * }
 */
func loadKnowledge() (*SystemKnowledge, error) {
	fname := filepath.Join(getCostrictDir(), "share", ".well-known.json")

	bytes, err := os.ReadFile(fname)
	if err != nil {
		return nil, fmt.Errorf("load 'completion-agent.json' failed: %v", err)
	}
	var c SystemKnowledge
	if err := json.Unmarshal(bytes, &c); err != nil {
		return nil, fmt.Errorf("unmarshal 'completion-agent.json' failed: %v", err)
	}
	return &c, nil
}

/**
 * Get codebase indexer service URL
 * @returns {string} Returns codebase service URL or default localhost URL
 * @description
 * - Loads system knowledge configuration
 * - Searches for codebase-indexer service in knowledge
 * - Returns configured service URL or default localhost:9001
 * - Used to construct codebase indexing API endpoints
 * @example
 * url := getCodebaseUrl()
 * // url will be "http://localhost:9001" or configured value
 */
func getCodebaseUrl() string {
	known, err := loadKnowledge()
	if err != nil {
		return "http://localhost:9001"
	}
	for _, s := range known.Services {
		if s.Name == "codebase-indexer" {
			return fmt.Sprintf("http://localhost:%d", s.Port)
		}
	}
	return "http://localhost:9001"
}

/**
 * Localize configuration by processing template strings
 * @param {SoftwareConfig} cfg - Software configuration to localize
 * @description
 * - Processes tokenizer path template in wrapper configuration
 * - Localizes context URLs (definition, relation, semantic)
 * - Processes model authorization and completion URL templates
 * - Applies environment-specific values to template strings
 * @example
 * localize(config)
 * // config fields will be updated with localized values
 */
func localize(cfg *SoftwareConfig) {
	cfg.Wrapper.Tokenizer.Path = localizeString(cfg.Wrapper.Tokenizer.Path)
	cfg.Context.Definition.Url = localizeString(cfg.Context.Definition.Url)
	cfg.Context.Relation.Url = localizeString(cfg.Context.Relation.Url)
	cfg.Context.Semantic.Url = localizeString(cfg.Context.Semantic.Url)
	for i, c := range cfg.Models {
		cfg.Models[i].Authorization = localizeString(c.Authorization)
		cfg.Models[i].CompletionsUrl = localizeString(c.CompletionsUrl)
	}
}

/**
 * Localize a template string using global environment variables
 * @param {string} s - Template string to localize
 * @returns {string} Returns localized string or original string if parsing fails
 * @description
 * - Parses input string as Go template
 * - Executes template with global environment variables
 * - Supports environment and authentication variables
 * - Logs fatal error and returns original string if template processing fails
 * @example
 * localized := localizeString("{{.Env.CostrictDir}}/path")
 * // localized will be "/actual/path/path"
 */
func localizeString(s string) string {
	tpl, err := template.New("config").Parse(s)
	if err != nil {
		logger.Fatal("failed to parse template", zap.Error(err))
		return s
	}

	var sBuf bytes.Buffer
	if err := tpl.Execute(&sBuf, globalEnv); err != nil {
		logger.Fatal("failed to execute template", zap.Error(err))
		return s
	}

	return sBuf.String()
}
