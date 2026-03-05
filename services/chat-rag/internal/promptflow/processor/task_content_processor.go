package processor

import (
	"fmt"
	"strings"

	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/model"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"go.uber.org/zap"
)

// TaskContentProcessor replaces content within <task>...</task> tags based on configuration
type TaskContentProcessor struct {
	BaseProcessor
	config     *config.PreciseContextConfig
	agentName  string
	promptMode string
}

// NewTaskContentProcessor creates a new TaskContentReplacer processor
func NewTaskContentProcessor(
	config *config.PreciseContextConfig,
	agentName string,
	promptMode string,
) *TaskContentProcessor {
	return &TaskContentProcessor{
		config:     config,
		agentName:  agentName,
		promptMode: promptMode,
	}
}

// Execute processes the prompt message to replace task content
func (t *TaskContentProcessor) Execute(promptMsg *PromptMsg) {
	logger.Info("Executing TaskContentReplacer",
		zap.String("agent", t.agentName),
		zap.String("prompt_mode", t.promptMode))

	// Find applicable rules for current agent and mode
	applicableRuleKeys := t.findApplicableRules()

	// Process message if any rules are applicable
	if len(applicableRuleKeys) > 0 {
		err := t.updateMessageContent(promptMsg, applicableRuleKeys)
		if err != nil {
			logger.Error("Failed to update message content", zap.Error(err))
		} else {
			logger.Info("Successfully processed task content")
		}
	}

	t.passToNext(promptMsg)
}

// findApplicableRules finds all rules that apply to the current agent and mode
func (t *TaskContentProcessor) findApplicableRules() []string {
	var applicableRuleKeys []string
	for ruleName, ruleConfig := range t.config.TaskContentReplaceRule {
		if t.isRuleApplicable(ruleConfig) {
			applicableRuleKeys = append(applicableRuleKeys, ruleName)
		}
	}

	logger.Info("Found applicable rules", zap.Strings("rules", applicableRuleKeys))
	return applicableRuleKeys
}

// isRuleApplicable checks if a rule applies to the current agent and mode
func (t *TaskContentProcessor) isRuleApplicable(ruleConfig config.TaskContentReplaceConfig) bool {
	// If no valid agents specified, rule applies to all
	if len(ruleConfig.ValidAgents) == 0 {
		return true
	}

	// Check if current mode is in the valid agents
	modeAgents, exists := ruleConfig.ValidAgents[t.promptMode]
	if !exists {
		return false
	}

	// Check if current agent is in the list for this mode
	for _, agent := range modeAgents {
		if agent == t.agentName {
			return true
		}
	}

	return false
}

// applyReplacements applies key-value replacements to the content
func (t *TaskContentProcessor) applyReplacements(content string, matchKeys map[string]string) (string, bool) {
	changed := false

	// Apply all replacements to the entire content
	// Use strings.NewReplacer to perform a single-pass replacement
	// This avoids recursive replacements where a replacement value triggers another rule
	var oldnew []string
	for key, value := range matchKeys {
		oldnew = append(oldnew, key, value)
	}

	if len(oldnew) > 0 {
		replacer := strings.NewReplacer(oldnew...)
		newContent := replacer.Replace(content)
		if newContent != content {
			content = newContent
			changed = true
		}
	}

	return content, changed
}

// updateMessageContent updates the content of a message with applicable replacement rules
func (t *TaskContentProcessor) updateMessageContent(promptMsg *PromptMsg, applicableRuleKeys []string) error {
	var msg *types.Message

	// Check if there are any older user messages, first chat has no older user messages
	if len(promptMsg.olderUserMsgList) == 0 {
		logger.Info("No older user messages found, using lastUserMsg instead")
		msg = promptMsg.lastUserMsg
	} else {
		// Get the first older user message
		msg = &promptMsg.olderUserMsgList[0]
	}

	// Use ExtractMsgContent to normalize content to []Content
	var contentExtractor model.Content
	contents, err := contentExtractor.ExtractMsgContent(msg)
	if err != nil {
		return fmt.Errorf("failed to extract message content: %w", err)
	}

	// Use the first content directly
	if len(contents) == 0 {
		logger.Info("No content found in message")
		return nil
	}

	// Process each applicable rule
	modifiedContent := contents[0].Text

	for _, ruleKey := range applicableRuleKeys {
		// Get rule config from TaskContentReplaceRule
		ruleConfig, exists := t.config.TaskContentReplaceRule[ruleKey]
		if !exists {
			logger.Warn("Rule not found in TaskContentReplaceRule",
				zap.String("rule", ruleKey))
			continue
		}

		// Check if content contains skip key
		if ruleConfig.SkipKey != "" && strings.Contains(modifiedContent, ruleConfig.SkipKey) {
			logger.Info("Content contains skip key, skipping rule",
				zap.String("rule", ruleKey),
				zap.String("skip_key", ruleConfig.SkipKey))
			continue
		}

		// Apply replacements
		newContent, changed := t.applyReplacements(modifiedContent, ruleConfig.MatchKeys)
		if !changed {
			logger.Info("No content changes applied for rule",
				zap.String("rule", ruleKey))
			continue
		}

		modifiedContent = newContent
		contents[0].Text = modifiedContent
		msg.Content = contents
		logger.Info("Applied task content replacements",
			zap.String("rule", ruleKey))
	}

	return nil
}
