package processor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/functions"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"go.uber.org/zap"
)

type XmlToolAdapter struct {
	BaseProcessor

	ctx          context.Context
	toolExecutor functions.ToolExecutor
	toolConfig   *config.ToolConfig
	agentName    string
	promptMode   string
}

func NewXmlToolAdapter(ctx context.Context, toolExecutor functions.ToolExecutor, toolConfig *config.ToolConfig, agentName string, promptMode string) *XmlToolAdapter {
	return &XmlToolAdapter{
		ctx:          ctx,
		toolExecutor: toolExecutor,
		toolConfig:   toolConfig,
		agentName:    agentName,
		promptMode:   promptMode,
	}
}

func (x *XmlToolAdapter) Execute(promptMsg *PromptMsg) {
	const method = "XmlToolAdapter.Execute"

	if promptMsg == nil {
		x.Err = fmt.Errorf("received prompt message is empty")
		logger.Error(x.Err.Error(), zap.String("method", method))
		return
	}

	// Check if all tools are disabled globally
	if x.toolConfig != nil && x.toolConfig.DisableTools {
		logger.InfoC(x.ctx, "All tools are disabled globally", zap.String("method", method))
		x.passToNext(promptMsg)
		return
	}

	systemContent, err := x.extractSystemContent(promptMsg.systemMsg)
	if err != nil {
		logger.WarnC(x.ctx, "Failed to extract system message content",
			zap.String("method", method),
			zap.Error(err))
		x.Err = fmt.Errorf("failed to extract system message content: %w", err)
		x.passToNext(promptMsg)
		return
	}

	// Check if this agent is disabled from using tools
	if x.toolConfig != nil && x.isAgentDisabled(x.agentName, x.promptMode) {
		logger.InfoC(x.ctx, "Agent is disabled from using tools",
			zap.String("agent", x.agentName), zap.String("mode", x.promptMode), zap.String("method", method))
		x.passToNext(promptMsg)
		return
	}

	// Process system content to insert tools
	updatedContent, err := x.insertToolsIntoSystemContent(systemContent)
	if err != nil {
		logger.WarnC(x.ctx, "Failed to insert tools into system content",
			zap.String("method", method),
			zap.Error(err))
		x.Err = fmt.Errorf("failed to insert tools into system content: %w", err)
		x.passToNext(promptMsg)
		return
	}

	// Update the system message with the modified content
	promptMsg.UpdateSystemMsg(updatedContent)

	x.Handled = true
	x.passToNext(promptMsg)
}

// insertToolsIntoSystemContent inserts tool descriptions under the "# Tools" section
func (x *XmlToolAdapter) insertToolsIntoSystemContent(content string) (string, error) {
	const method = "XmlToolAdapter.insertToolsIntoSystemContent"

	// Add nil check to prevent panic
	if x.toolExecutor == nil {
		logger.ErrorC(x.ctx, "ToolExecutor is nil, skipping tool adaptation", zap.String("method", method))
		return content, nil
	}

	if len(x.toolExecutor.GetAllTools()) == 0 {
		logger.InfoC(x.ctx, "No tools available", zap.String("method", method))
	}

	// Combine all tools into a single string
	var toolsContent strings.Builder
	var capabilitiesContent strings.Builder
	var ruleContent strings.Builder

	toolNames := x.toolExecutor.GetAllTools()
	if len(toolNames) == 0 {
		logger.InfoC(x.ctx, "No tools available", zap.String("method", method))
	}

	// Parallel processing of tool checks and description retrieval
	type toolResult struct {
		name       string
		ready      bool
		readyErr   error
		desc       string
		descErr    error
		capability string
		capErr     error
		rule       string
		ruleErr    error
	}

	results := make([]toolResult, len(toolNames))
	var wg sync.WaitGroup

	for i, toolName := range toolNames {
		wg.Add(1)
		go func(index int, name string) {
			defer wg.Done()

			result := toolResult{name: name}

			// Check if tool is ready
			result.ready, result.readyErr = x.toolExecutor.CheckToolReady(x.ctx, name)

			if result.ready {
				// Get tool description
				result.desc, result.descErr = x.toolExecutor.GetToolDescription(name)

				// Get tool capability
				result.capability, result.capErr = x.toolExecutor.GetToolCapability(name)

				// Get tool rules
				result.rule, result.ruleErr = x.toolExecutor.GetToolRule(name)
			}

			results[index] = result
		}(i, toolName)
	}

	wg.Wait()

	// Process results and build content
	for _, result := range results {
		if !result.ready {
			logger.WarnC(x.ctx, "Tool is not ready, skip adapt", zap.String("tool", result.name),
				zap.String("method", method), zap.Error(result.readyErr))
			continue
		}

		if result.descErr != nil {
			if errors.Is(result.descErr, context.Canceled) || errors.Is(x.ctx.Err(), context.Canceled) {
				logger.WarnC(x.ctx, "Context canceled getting tool description", zap.String("tool", result.name), zap.Error(result.descErr))
			} else {
				logger.Error("Failed to get tool description", zap.Error(result.descErr))
			}
			continue
		}

		toolsContent.WriteString(result.desc)
		toolsContent.WriteString("\n\n")

		if result.capErr != nil {
			if errors.Is(result.capErr, context.Canceled) || errors.Is(x.ctx.Err(), context.Canceled) {
				logger.WarnC(x.ctx, "Context canceled getting tool capability", zap.String("tool", result.name), zap.Error(result.capErr))
			} else {
				logger.Error("Failed to get tool capability", zap.Error(result.capErr))
			}
			continue
		}
		capabilitiesContent.WriteString(result.capability)

		// Collect rules from ready tools
		if result.ruleErr != nil {
			if errors.Is(result.ruleErr, context.Canceled) || errors.Is(x.ctx.Err(), context.Canceled) {
				logger.WarnC(x.ctx, "Context canceled getting tool rules", zap.String("tool", result.name), zap.Error(result.ruleErr))
			} else {
				logger.Error("Failed to get tool rules", zap.Error(result.ruleErr))
			}
			continue
		}
		if result.rule != "" {
			ruleContent.WriteString(result.rule)
		}

		logger.InfoC(x.ctx, "Tool adapted in system prompt", zap.String("name", result.name))
	}

	// Insert the tools content after the tools header
	result, err := insertContentAfterMarker(content, "# Tools", toolsContent.String())
	if err != nil {
		return content, fmt.Errorf("failed to insert tools content: %w", err)
	}

	// Insert tool capabilities after CAPABILITIES section
	result, err = insertContentAfterMarker(result, "\n\n====\n\nCAPABILITIES\n\n", capabilitiesContent.String())
	if err != nil {
		return result, fmt.Errorf("failed to insert capabilities content: %w", err)
	}

	// Insert tools rules after RULES section
	result, err = insertContentAfterMarker(result, "\n\n====\n\nRULES\n\n", ruleContent.String())
	if err != nil {
		return result, fmt.Errorf("failed to insert rules content: %w", err)
	}

	return result, nil
}

// insertContentAfterMarker inserts content after a specific marker in the text
func insertContentAfterMarker(content, marker, newContent string) (string, error) {
	markerIndex := strings.Index(content, marker)
	if markerIndex == -1 {
		return content, fmt.Errorf("marker not found in content")
	}

	// For headers like "# Tools", find the end of the line
	if strings.HasPrefix(marker, "#") {
		lineEnd := strings.Index(content[markerIndex:], "\n")
		if lineEnd == -1 {
			lineEnd = len(content) - markerIndex
		}
		insertPos := markerIndex + lineEnd + 1
		return content[:insertPos] + "\n" + newContent + content[insertPos:], nil
	}

	// For other markers, insert immediately after the marker
	insertPos := markerIndex + len(marker)
	return content[:insertPos] + newContent + content[insertPos:], nil
}

// isAgentDisabled checks if the agent is disabled from using tools in the current mode
func (x *XmlToolAdapter) isAgentDisabled(agentName, mode string) bool {
	if x.toolConfig == nil || x.toolConfig.DisabledAgents == nil {
		return false
	}

	// Check if the mode exists in the disabled agents configuration
	disabledAgents, exists := x.toolConfig.DisabledAgents[mode]
	if !exists {
		return false
	}

	// Check if the agent is in the disabled list for this mode
	for _, disabledAgent := range disabledAgents {
		if disabledAgent == agentName {
			return true
		}
	}

	return false
}
