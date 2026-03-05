package processor

import (
	"fmt"
	"strings"

	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/tokenizer"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"github.com/zgsm-ai/chat-rag/internal/utils"
	"go.uber.org/zap"
)

type UserMsgFilter struct {
	BaseProcessor

	preciseContextConfig *config.PreciseContextConfig
	promptMode           string
	agentName            string
	tokenCounter         *tokenizer.TokenCounter
	TokenMetrics         types.TokenMetrics
}

func NewUserMsgFilter(
	preciseContextConfig *config.PreciseContextConfig,
	promptMode, agentName string,
	tokenCounter *tokenizer.TokenCounter,
) *UserMsgFilter {
	return &UserMsgFilter{
		preciseContextConfig: preciseContextConfig,
		promptMode:           promptMode,
		agentName:            agentName,
		tokenCounter:         tokenCounter,
	}
}

func (u *UserMsgFilter) Execute(promptMsg *PromptMsg) {
	const method = "UserMsgFilter.Execute"

	if promptMsg == nil {
		u.Err = fmt.Errorf("received prompt message is empty")
		logger.Error(u.Err.Error(), zap.String("method", method))
		return
	}

	// Calculate original token counts before filtering
	u.calculateTokenStats(promptMsg, true)

	// u.filterDuplicateMessages(promptMsg)

	u.filterAssistantToolPatterns(promptMsg)

	u.filterEnvironmentDetails(promptMsg)

	u.filterModesChangeContent(promptMsg)

	// Calculate processed token counts after filtering
	u.calculateTokenStats(promptMsg, false)

	// Calculate ratios and log metrics
	if u.tokenCounter != nil {
		u.TokenMetrics.CalculateRatios()

		logger.Info("Token metrics calculated",
			zap.Int("original_tokens", u.TokenMetrics.Original.All),
			zap.Int("processed_tokens", u.TokenMetrics.Processed.All),
			zap.Float64("token_ratio", u.TokenMetrics.Ratios.AllRatio),
			zap.String("method", method))
	}

	u.Handled = true
	u.passToNext(promptMsg)
}

// filterDuplicateMessages removes duplicate string content messages, keeping the last occurrence
func (u *UserMsgFilter) filterDuplicateMessages(promptMsg *PromptMsg) {
	const method = "UserMsgFilter.filterDuplicateMessages"

	// Skip processing if there are no older messages
	if len(promptMsg.olderUserMsgList) == 0 {
		return
	}

	originalCount := len(promptMsg.olderUserMsgList)
	seenContents := make(map[string]struct{})
	filteredMessages := make([]types.Message, 0, len(promptMsg.olderUserMsgList))

	// Iterate in reverse to keep the last occurrence of each duplicate
	for i := len(promptMsg.olderUserMsgList) - 1; i >= 0; i-- {
		msg := promptMsg.olderUserMsgList[i]

		content, ok := msg.Content.(string)
		if !ok {
			// Include non-string content messages as-is
			filteredMessages = append(filteredMessages, msg)
			continue
		}

		// Skip if we've already seen this content
		if _, exists := seenContents[content]; exists {
			continue
		}

		// Mark content as seen and add to filtered list
		seenContents[content] = struct{}{}
		filteredMessages = append(filteredMessages, msg)
	}

	// Reverse back to original order (now with duplicates removed)
	for i, j := 0, len(filteredMessages)-1; i < j; i, j = i+1, j-1 {
		filteredMessages[i], filteredMessages[j] = filteredMessages[j], filteredMessages[i]
	}

	promptMsg.olderUserMsgList = filteredMessages

	removedCount := originalCount - len(promptMsg.olderUserMsgList)
	logger.Info("removed duplicate content count",
		zap.Int("removedCount", removedCount),
		zap.String("method", method))
}

// TODO this func will be removed when client apapted tool status dispply
// filterAssistantToolPatterns removes tool execution patterns from assistant messages
func (u *UserMsgFilter) filterAssistantToolPatterns(promptMsg *PromptMsg) {
	for i := range promptMsg.olderUserMsgList {
		msg := &promptMsg.olderUserMsgList[i]

		// Only process assistant messages
		if msg.Role != types.RoleAssistant {
			continue
		}

		content, ok := msg.Content.(string)
		if !ok {
			// Skip non-string content messages
			continue
		}

		// Remove tool execution patterns
		msg.Content = u.removeToolExecutionPatterns(content)
	}
}

// removeToolExecutionPatterns removes strings that executing tool
// Temporarily hardcoded
func (u *UserMsgFilter) removeToolExecutionPatterns(content string) string {
	startPattern := types.StrFilterToolSearchStart
	endPattern := types.StrFilterToolSearchEnd + "....."

	result := content
	for {
		startIndex := u.indexOf(result, startPattern)
		if startIndex == -1 {
			break
		}

		endIndex := u.indexOf(result[startIndex:], endPattern)
		if endIndex == -1 {
			break
		}

		endIndex += startIndex // Adjust to original string index
		// Include the end pattern length
		endIndex += len(endPattern)

		// Remove the pattern
		result = result[:startIndex] + result[endIndex:]
		logger.Info("removed tool executing... content", zap.String("method", "removeToolExecutionPatterns"))
	}

	// Remove the specific string
	thinkPattern := types.StrFilterToolAnalyzing + "..."
	result = strings.ReplaceAll(result, thinkPattern, "")
	if result != content {
		logger.Info("removed thinking... content", zap.String("method", "removeToolExecutionPatterns"))
	}

	return result
}

// indexOf returns the index of the first occurrence of pattern in s
func (u *UserMsgFilter) indexOf(s, pattern string) int {
	for i := 0; i <= len(s)-len(pattern); i++ {
		if s[i:i+len(pattern)] == pattern {
			return i
		}
	}
	return -1
}

// filterEnvironmentDetails removes environment details content from user messages
// Keeps the first occurrence of <environment_details> and removes subsequent ones
func (u *UserMsgFilter) filterEnvironmentDetails(promptMsg *PromptMsg) {
	const method = "UserMsgFilter.filterEnvironmentDetails"
	const environment_details = "<environment_details>"

	// Check if environment details filter is enabled
	if !u.preciseContextConfig.EnableEnvDetailsFilter {
		logger.Info("environment details filter is disabled, skipping", zap.String("method", method))
		return
	}

	removedCount := 0
	environmentDetailsCount := 0

	for i := range promptMsg.olderUserMsgList {
		msg := &promptMsg.olderUserMsgList[i]

		// Only process user messages
		if msg.Role != types.RoleUser {
			continue
		}

		// Check if msg.Content is a list
		contentList, ok := msg.Content.([]interface{})
		if !ok {
			continue
		}

		// Filter out environment details from content list in place
		for j := 0; j < len(contentList); {
			item := contentList[j]

			// Try to extract text from the content item
			textStr := utils.ExtractTextFromContent(item)
			if textStr == "" {
				j++
				continue
			}

			// Check if this is environment details
			if strings.HasPrefix(textStr, environment_details) {
				environmentDetailsCount++
				if environmentDetailsCount > 1 {
					// Remove this element by slicing it out (skip the first one)
					contentList = append(contentList[:j], contentList[j+1:]...)
					removedCount++
					// Don't increment j since we removed the current element
					continue
				}
			}

			// Move to next element
			j++
		}

		// Update msg.Content with the modified slice
		msg.Content = contentList

	}

	logger.Info("[environment details] filtering completed",
		zap.Int("removed_count", removedCount),
		zap.String("method", method))
}

// filterModesChangeContent removes ModesChange related content from system message
// based on DisabledModesChangeAgents configuration
func (u *UserMsgFilter) filterModesChangeContent(promptMsg *PromptMsg) {
	const method = "UserMsgFilter.filterModesChangeContent"

	// Check if DisabledModesChangeAgents is configured
	if u.preciseContextConfig.DisabledModesChangeAgents == nil {
		return
	}

	// Check if current prompt mode is in the disabled configuration
	agents, exists := u.preciseContextConfig.DisabledModesChangeAgents[u.promptMode]
	if !exists {
		return
	}

	// Check if current agent is in the disabled list for this mode
	shouldFilter := false
	for _, agent := range agents {
		if agent == u.agentName {
			shouldFilter = true
			break
		}
	}

	if !shouldFilter {
		return
	}

	// Filter ModesChange content from system message
	if promptMsg.systemMsg == nil {
		return
	}

	systemContent, err := u.extractSystemContent(promptMsg.systemMsg)
	if err != nil {
		logger.Warn("Failed to extract system message content for ModesChange filtering",
			zap.String("method", method),
			zap.Error(err))
		return
	}

	// Remove content sections using helper function
	result := systemContent
	result = u.removeContentSection(result, "## switch_mode", "\n##", "switch_mode", method)
	result = u.removeContentSection(result, "====\n\nMODES", "\n====", "modes_desc", method)

	// Update system message content
	promptMsg.UpdateSystemMsg(result)
}

// removeContentSection removes a section of content between startPattern and endPattern
// Returns the modified content string
func (u *UserMsgFilter) removeContentSection(content, startPattern, endPattern, sectionName, method string) string {
	startIndex := strings.Index(content, startPattern)
	if startIndex == -1 {
		return content
	}

	// Find the end of the section
	endIndex := strings.Index(content[startIndex:], endPattern)
	if endIndex == -1 {
		return content
	}

	endIndex += startIndex // Adjust to original string index

	// Remove the section
	result := content[:startIndex] + content[endIndex:]
	logger.Info(fmt.Sprintf("removed %s content from system message", sectionName),
		zap.String("prompt_mode", u.promptMode),
		zap.String("agent", u.agentName),
		zap.String("method", method))

	return result
}

// calculateTokenStats calculates token statistics for the given prompt message
// isOriginal indicates whether to calculate for original (true) or processed (false) state
func (u *UserMsgFilter) calculateTokenStats(promptMsg *PromptMsg, isOriginal bool) {
	if u.tokenCounter == nil {
		return
	}

	// Count tokens for older user messages
	userTokens := u.tokenCounter.CountMessagesTokens(promptMsg.olderUserMsgList)

	// Count tokens for system message if exists
	systemTokens := 0
	if promptMsg.systemMsg != nil {
		systemTokens = u.tokenCounter.CountOneMessageTokens(*promptMsg.systemMsg)
	}

	// Create token stats
	tokenStats := types.TokenStats{
		SystemTokens: systemTokens,
		UserTokens:   userTokens,
		All:          systemTokens + userTokens,
	}

	// Set to appropriate field based on isOriginal flag
	if isOriginal {
		u.TokenMetrics.Original = tokenStats
	} else {
		u.TokenMetrics.Processed = tokenStats
	}
}
